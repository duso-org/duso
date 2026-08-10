package runtime

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	"github.com/duso-org/duso/pkg/script"
)

// builtinSendMail sends a single email through an SMTP relay.
//
// duso never receives mail and never acts as a mail server; this opens an
// outbound submission connection, hands over one message, and is done.
//
// Options (one flat object):
//   - host (string, required)  - relay hostname
//   - port (number)            - default 587
//   - user, pass (string)      - AUTH credentials, omit for unauthenticated relays
//   - tls (string)             - "auto" (default), "implicit", "starttls", "none"
//   - timeout (number)         - seconds, default 30
//   - from (string, required)  - envelope sender, may be "Name <a@b.com>"
//   - to (string or array, required)
//   - cc, bcc (string or array)
//   - subject (string)
//   - text (string)            - text/plain body
//   - html (string)            - text/html body
//   - headers (object)         - extra headers, e.g. Reply-To
//
// Returns {ok = true, message_id = "<...>"}.
//
// Example:
//
//	send_mail({
//	  host = "smtp-relay.brevo.com", user = "u", pass = env("SMTP_PASS"),
//	  from = "app@example.com", to = "you@example.com",
//	  subject = "Password reset", text = "Click the link."
//	})
func builtinSendMail(evaluator *Evaluator, args map[string]any) (any, error) {
	opts, err := smtpOptions(args)
	if err != nil {
		return nil, err
	}

	host := smtpString(opts, "host", "")
	if host == "" {
		return nil, errors.New("send_mail() requires a host")
	}
	port := int(smtpFloat(opts, "port", 587))

	from := smtpString(opts, "from", "")
	if from == "" {
		return nil, errors.New("send_mail() requires a from address")
	}
	fromAddr, err := smtpBareAddress(from)
	if err != nil {
		return nil, fmt.Errorf("send_mail() invalid from address: %w", err)
	}

	to := smtpAddressList(opts["to"])
	cc := smtpAddressList(opts["cc"])
	bcc := smtpAddressList(opts["bcc"])
	if len(to) == 0 {
		return nil, errors.New("send_mail() requires at least one to address")
	}

	// Envelope recipients include bcc; the message headers never will.
	var rcpts []string
	for _, list := range [][]string{to, cc, bcc} {
		for _, a := range list {
			bare, err := smtpBareAddress(a)
			if err != nil {
				return nil, fmt.Errorf("send_mail() invalid recipient %q: %w", a, err)
			}
			rcpts = append(rcpts, bare)
		}
	}

	msg, messageID, err := buildMailMessage(opts, from, to, cc)
	if err != nil {
		return nil, err
	}

	// Derive from the spawned process context so kill(pid) aborts an in-flight
	// send, matching fetch()'s behavior.
	parent := context.Background()
	if reqCtx, ok := script.CurrentRequestContext(evaluator); ok && reqCtx.ProcessCtx != nil {
		parent = reqCtx.ProcessCtx
	}
	timeout := smtpFloat(opts, "timeout", 30)
	if timeout <= 0 {
		timeout = 30
	}
	ctx, cancel := context.WithTimeout(parent, time.Duration(timeout*float64(time.Second)))
	defer cancel()

	cfg := smtpDialConfig{
		host:     host,
		port:     port,
		user:     smtpString(opts, "user", ""),
		pass:     smtpString(opts, "pass", ""),
		tlsMode:  strings.ToLower(smtpString(opts, "tls", "auto")),
		insecure: smtpBool(opts, "insecure"),
	}

	if err := smtpDeliver(ctx, cfg, fromAddr, rcpts, msg); err != nil {
		return nil, fmt.Errorf("send_mail() failed: %w", err)
	}

	return map[string]any{
		"ok":         true,
		"message_id": messageID,
	}, nil
}

// smtpOptions pulls the single options object from positional or named args.
func smtpOptions(args map[string]any) (map[string]any, error) {
	for _, key := range []string{"0", "options"} {
		if v, ok := args[key]; ok {
			if m, ok := v.(map[string]any); ok {
				return m, nil
			}
			return nil, errors.New("send_mail() options must be an object")
		}
	}
	// Allow named arguments spread directly: send_mail(host = ..., to = ...)
	if len(args) > 0 {
		return args, nil
	}
	return nil, errors.New("send_mail() requires an options object")
}

// smtpScalar renders a single value as a string, unwrapping the script-level
// representations that reach a builtin (Value, ValueRef) as well as plain Go types.
func smtpScalar(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case script.Value:
		return t.String()
	case *script.ValueRef:
		return t.Val.String()
	default:
		return fmt.Sprintf("%v", t)
	}
}

// smtpString reads a string option, defaulting when absent or empty.
func smtpString(opts map[string]any, key, defaultVal string) string {
	v, ok := opts[key]
	if !ok || v == nil {
		return defaultVal
	}
	if s := smtpScalar(v); s != "" {
		return s
	}
	return defaultVal
}

// smtpFloat reads a numeric option, defaulting when absent or unparseable.
func smtpFloat(opts map[string]any, key string, defaultVal float64) float64 {
	v, ok := opts[key]
	if !ok || v == nil {
		return defaultVal
	}
	switch t := v.(type) {
	case float64:
		return t
	case int:
		return float64(t)
	case script.Value:
		if t.Type == script.VAL_NUMBER {
			return t.AsNumber()
		}
	case *script.ValueRef:
		if t.Val.Type == script.VAL_NUMBER {
			return t.Val.AsNumber()
		}
	}
	if f, err := strconv.ParseFloat(strings.TrimSpace(smtpScalar(v)), 64); err == nil {
		return f
	}
	return defaultVal
}

// smtpBool reads a boolean option.
func smtpBool(opts map[string]any, key string) bool {
	switch t := opts[key].(type) {
	case bool:
		return t
	case script.Value:
		return t.AsBool()
	case *script.ValueRef:
		return t.Val.AsBool()
	}
	return false
}

// smtpAddressList normalizes a string or array of strings into a slice.
// Duso arrays reach builtins as *[]script.Value; plain []any is accepted too.
func smtpAddressList(v any) []string {
	appendNonEmpty := func(out []string, s string) []string {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
		return out
	}

	switch t := v.(type) {
	case nil:
		return nil
	case *script.ValueRef:
		return smtpAddressList(t.Val)
	case script.Value:
		if arr := t.AsArray(); arr != nil {
			var out []string
			for _, item := range arr {
				out = appendNonEmpty(out, item.String())
			}
			return out
		}
		return appendNonEmpty(nil, t.String())
	case *[]script.Value:
		var out []string
		for _, item := range *t {
			out = appendNonEmpty(out, item.String())
		}
		return out
	case []any:
		var out []string
		for _, item := range t {
			out = appendNonEmpty(out, smtpScalar(item))
		}
		return out
	default:
		return appendNonEmpty(nil, smtpScalar(t))
	}
}

// smtpObject unwraps a nested object argument into a plain map.
func smtpObject(v any) map[string]any {
	switch t := v.(type) {
	case map[string]any:
		return t
	case *script.ValueRef:
		return smtpObject(t.Val)
	case script.Value:
		if obj := t.AsObject(); obj != nil {
			out := make(map[string]any, len(obj))
			for k, val := range obj {
				out[k] = val
			}
			return out
		}
	case map[string]script.Value:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[k] = val
		}
		return out
	}
	return nil
}

// smtpBareAddress strips any display name, leaving just user@host for the envelope.
func smtpBareAddress(s string) (string, error) {
	trimmed := strings.TrimSpace(s)
	// Reject control characters up front: a CRLF here would forge an SMTP
	// command or an extra header, whichever consumed the value first.
	if strings.ContainsFunc(trimmed, func(r rune) bool { return r < ' ' || r == 0x7f }) {
		return "", errors.New("address contains control characters")
	}
	addr, err := mail.ParseAddress(trimmed)
	if err != nil {
		// Be lenient: a bare address without angle brackets is common and valid.
		if strings.Contains(trimmed, "@") && !strings.ContainsAny(trimmed, " <>,") {
			return trimmed, nil
		}
		return "", err
	}
	return addr.Address, nil
}

// smtpValidHeaderName reports whether a caller-supplied header name is a bare
// RFC 5322 token, so it cannot introduce a line break or a colon of its own.
func smtpValidHeaderName(name string) bool {
	if name == "" {
		return false
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		case r == '-', r == '_':
		default:
			return false
		}
	}
	return true
}

// smtpSanitizeHeaderValue folds line breaks to spaces so a value can never
// start a new header, and drops other control characters.
func smtpSanitizeHeaderValue(v string) string {
	return strings.TrimSpace(strings.Map(func(r rune) rune {
		switch {
		case r == '\r' || r == '\n' || r == '\t':
			return ' '
		case r < ' ' || r == 0x7f:
			return -1
		}
		return r
	}, v))
}

// buildMailMessage assembles RFC 5322 headers plus a single-part or
// multipart/alternative body. Returns the wire bytes and the Message-ID.
func buildMailMessage(opts map[string]any, from string, to, cc []string) ([]byte, string, error) {
	text := smtpString(opts, "text", "")
	html := smtpString(opts, "html", "")
	if text == "" && html == "" {
		return nil, "", errors.New("send_mail() requires a text or html body")
	}

	messageID, err := smtpMessageID(from)
	if err != nil {
		return nil, "", err
	}

	var buf bytes.Buffer
	header := func(name, value string) {
		fmt.Fprintf(&buf, "%s: %s\r\n", name, value)
	}

	header("From", smtpEncodeAddress(from))
	header("To", smtpEncodeAddressList(to))
	if len(cc) > 0 {
		header("Cc", smtpEncodeAddressList(cc))
	}
	// bcc is deliberately absent: it belongs to the envelope only.
	if subject := smtpString(opts, "subject", ""); subject != "" {
		header("Subject", mime.QEncoding.Encode("utf-8", subject))
	}
	header("Date", time.Now().Format(time.RFC1123Z))
	header("Message-ID", messageID)

	// Caller-supplied extras (Reply-To, List-Unsubscribe, ...) never override
	// the headers we control above.
	if extra := smtpObject(opts["headers"]); extra != nil {
		for k, v := range extra {
			switch textproto.CanonicalMIMEHeaderKey(k) {
			case "From", "To", "Cc", "Bcc", "Subject", "Date", "Message-Id",
				"Mime-Version", "Content-Type", "Content-Transfer-Encoding":
				continue
			}
			// A header name is a token; anything else could forge a new line.
			if !smtpValidHeaderName(k) {
				continue
			}
			header(k, smtpSanitizeHeaderValue(smtpScalar(v)))
		}
	}

	header("MIME-Version", "1.0")

	// Only wrap in multipart when there is genuinely more than one part; a
	// single part inside a multipart shell reads as machine-generated.
	if text != "" && html != "" {
		mp := multipart.NewWriter(&buf)
		header("Content-Type", "multipart/alternative; boundary=\""+mp.Boundary()+"\"")
		buf.WriteString("\r\n")

		// Least-rich part first, per RFC 2046.
		if err := writeMailPart(mp, "text/plain; charset=utf-8", text); err != nil {
			return nil, "", err
		}
		if err := writeMailPart(mp, "text/html; charset=utf-8", html); err != nil {
			return nil, "", err
		}
		if err := mp.Close(); err != nil {
			return nil, "", err
		}
		return buf.Bytes(), messageID, nil
	}

	body := text
	contentType := "text/plain; charset=utf-8"
	if text == "" {
		body = html
		contentType = "text/html; charset=utf-8"
	}
	header("Content-Type", contentType)
	header("Content-Transfer-Encoding", "quoted-printable")
	buf.WriteString("\r\n")
	if err := writeQuotedPrintable(&buf, body); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), messageID, nil
}

// writeMailPart adds one quoted-printable part to a multipart body.
func writeMailPart(mp *multipart.Writer, contentType, body string) error {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Type", contentType)
	h.Set("Content-Transfer-Encoding", "quoted-printable")
	w, err := mp.CreatePart(h)
	if err != nil {
		return err
	}
	return writeQuotedPrintable(w, body)
}

// writeQuotedPrintable encodes a body, normalizing line endings first so the
// encoder emits consistent CRLF and never exceeds SMTP's line length limit.
func writeQuotedPrintable(w interface{ Write([]byte) (int, error) }, body string) error {
	normalized := strings.ReplaceAll(strings.ReplaceAll(body, "\r\n", "\n"), "\r", "\n")
	qp := quotedprintable.NewWriter(w)
	if _, err := qp.Write([]byte(normalized)); err != nil {
		return err
	}
	return qp.Close()
}

// smtpEncodeAddress RFC 2047-encodes a display name while leaving the address alone.
func smtpEncodeAddress(s string) string {
	addr, err := mail.ParseAddress(strings.TrimSpace(s))
	if err != nil || addr.Name == "" {
		return strings.TrimSpace(s)
	}
	return (&mail.Address{Name: addr.Name, Address: addr.Address}).String()
}

func smtpEncodeAddressList(list []string) string {
	out := make([]string, 0, len(list))
	for _, a := range list {
		out = append(out, smtpEncodeAddress(a))
	}
	return strings.Join(out, ", ")
}

// smtpMessageID builds a globally unique Message-ID rooted at the sender's domain.
func smtpMessageID(from string) (string, error) {
	domain := "localhost"
	if bare, err := smtpBareAddress(from); err == nil {
		if i := strings.LastIndex(bare, "@"); i >= 0 {
			domain = bare[i+1:]
		}
	}
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return fmt.Sprintf("<%s@%s>", hex.EncodeToString(b[:]), domain), nil
}

type smtpDialConfig struct {
	host     string
	port     int
	user     string
	pass     string
	tlsMode  string
	insecure bool
}

func (c smtpDialConfig) addr() string {
	return net.JoinHostPort(c.host, fmt.Sprintf("%d", c.port))
}

// smtpDeliver opens a connection, sends one message, and closes it.
//
// Deliberately not pooled. Relays drop idle sessions within 30-60 seconds, so a
// cache would miss for any app that sends occasionally, while a shared connection
// would force concurrent senders to serialize on one SMTP session. Dialing per
// send keeps concurrent sends genuinely concurrent; the handshake cost is
// negligible beside the request that triggered the mail.
func smtpDeliver(ctx context.Context, cfg smtpDialConfig, from string, rcpts []string, msg []byte) error {
	client, err := smtpDial(ctx, cfg)
	if err != nil {
		return err
	}
	if err := smtpSendOn(ctx, client, from, rcpts, msg); err != nil {
		client.Close()
		return err
	}
	// Quit waits for the relay to confirm it queued the message; Close would not.
	return client.Quit()
}

// smtpDial opens a connection, negotiates TLS, and authenticates.
func smtpDial(ctx context.Context, cfg smtpDialConfig) (*smtp.Client, error) {
	tlsConfig := &tls.Config{
		ServerName:         cfg.host,
		InsecureSkipVerify: cfg.insecure,
	}

	mode := cfg.tlsMode
	if mode == "" || mode == "auto" {
		// 465 is implicit TLS by convention; everything else negotiates upward.
		if cfg.port == 465 {
			mode = "implicit"
		} else {
			mode = "starttls"
		}
	}

	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", cfg.addr())
	if err != nil {
		return nil, err
	}
	// Keep the whole session bounded by the caller's timeout, including reads.
	if deadline, ok := ctx.Deadline(); ok {
		conn.SetDeadline(deadline)
	}

	if mode == "implicit" {
		conn = tls.Client(conn, tlsConfig)
		if err := conn.(*tls.Conn).HandshakeContext(ctx); err != nil {
			conn.Close()
			return nil, err
		}
	}

	client, err := smtp.NewClient(conn, cfg.host)
	if err != nil {
		conn.Close()
		return nil, err
	}

	if mode == "starttls" {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(tlsConfig); err != nil {
				client.Close()
				return nil, err
			}
		}
	}

	if cfg.user != "" {
		auth, err := smtpAuth(client, cfg)
		if err != nil {
			client.Close()
			return nil, err
		}
		if err := client.Auth(auth); err != nil {
			client.Close()
			return nil, err
		}
	}
	return client, nil
}

// smtpAuth picks a mechanism the server actually advertises. PLAIN is the norm;
// LOGIN is not in net/smtp but some Exchange-flavored relays require it.
func smtpAuth(client *smtp.Client, cfg smtpDialConfig) (smtp.Auth, error) {
	ok, mechs := client.Extension("AUTH")
	if !ok {
		return nil, errors.New("server does not support authentication")
	}
	upper := strings.ToUpper(mechs)
	switch {
	case strings.Contains(upper, "PLAIN"):
		return smtp.PlainAuth("", cfg.user, cfg.pass, cfg.host), nil
	case strings.Contains(upper, "LOGIN"):
		return &smtpLoginAuth{user: cfg.user, pass: cfg.pass, host: cfg.host}, nil
	case strings.Contains(upper, "CRAM-MD5"):
		return smtp.CRAMMD5Auth(cfg.user, cfg.pass), nil
	}
	return nil, fmt.Errorf("no supported auth mechanism (server offers: %s)", mechs)
}

// smtpSendOn runs one MAIL/RCPT/DATA transaction on an existing client.
func smtpSendOn(ctx context.Context, client *smtp.Client, from string, rcpts []string, msg []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	for _, rcpt := range rcpts {
		if err := client.Rcpt(rcpt); err != nil {
			return err
		}
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		w.Close()
		return err
	}
	return w.Close()
}

// smtpLoginAuth implements the non-standard AUTH LOGIN mechanism.
type smtpLoginAuth struct {
	user, pass, host string
}

func (a *smtpLoginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	if !server.TLS && !smtpIsLocalhost(server.Name) {
		return "", nil, errors.New("smtp: refusing to send LOGIN credentials over an unencrypted connection")
	}
	return "LOGIN", nil, nil
}

func (a *smtpLoginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}
	switch strings.ToLower(strings.TrimRight(string(fromServer), ": ")) {
	case "username":
		return []byte(a.user), nil
	case "password":
		return []byte(a.pass), nil
	}
	return nil, fmt.Errorf("smtp: unexpected LOGIN challenge %q", fromServer)
}

func smtpIsLocalhost(name string) bool {
	return name == "localhost" || name == "127.0.0.1" || name == "::1"
}
