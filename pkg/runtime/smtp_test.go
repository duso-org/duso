package runtime

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"

	"github.com/duso-org/duso/pkg/script"
)

// fakeSession records one SMTP transaction as the server saw it.
type fakeSession struct {
	from   string
	rcpts  []string
	data   string
	mech   string
	authed bool
}

// fakeSMTP is a minimal in-process SMTP server: enough of RFC 5321 to accept a
// message and report exactly what arrived on the wire.
type fakeSMTP struct {
	ln        net.Listener
	mu        sync.Mutex
	sessions  []*fakeSession
	connCount int
	authMechs string
	wg        sync.WaitGroup
}

func newFakeSMTP(t *testing.T, authMechs string) *fakeSMTP {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	s := &fakeSMTP{ln: ln, authMechs: authMechs}
	s.wg.Add(1)
	go s.serve()
	t.Cleanup(func() {
		ln.Close()
		s.wg.Wait()
	})
	return s
}

func (s *fakeSMTP) serve() {
	defer s.wg.Done()
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return
		}
		s.mu.Lock()
		s.connCount++
		s.mu.Unlock()
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handle(conn)
		}()
	}
}

func (s *fakeSMTP) handle(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	say := func(format string, a ...any) {
		fmt.Fprintf(w, format+"\r\n", a...)
		w.Flush()
	}

	say("220 fake.test ESMTP")
	session := &fakeSession{}

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		verb := strings.ToUpper(line)

		switch {
		case strings.HasPrefix(verb, "EHLO"):
			say("250-fake.test")
			if s.authMechs != "" {
				say("250-AUTH %s", s.authMechs)
			}
			say("250 SIZE 10485760")

		case strings.HasPrefix(verb, "HELO"):
			say("250 fake.test")

		case strings.HasPrefix(verb, "AUTH PLAIN"):
			session.mech, session.authed = "PLAIN", true
			say("235 2.7.0 authenticated")

		case strings.HasPrefix(verb, "AUTH LOGIN"):
			session.mech = "LOGIN"
			say("334 %s", base64.StdEncoding.EncodeToString([]byte("Username:")))
			if _, err := r.ReadString('\n'); err != nil {
				return
			}
			say("334 %s", base64.StdEncoding.EncodeToString([]byte("Password:")))
			if _, err := r.ReadString('\n'); err != nil {
				return
			}
			session.authed = true
			say("235 2.7.0 authenticated")

		case strings.HasPrefix(verb, "MAIL FROM:"):
			session.from = smtpTestAngleAddr(line[len("MAIL FROM:"):])
			say("250 2.1.0 ok")

		case strings.HasPrefix(verb, "RCPT TO:"):
			session.rcpts = append(session.rcpts, smtpTestAngleAddr(line[len("RCPT TO:"):]))
			say("250 2.1.5 ok")

		case verb == "DATA":
			say("354 end with <CRLF>.<CRLF>")
			var body strings.Builder
			for {
				dl, err := r.ReadString('\n')
				if err != nil {
					return
				}
				if dl == ".\r\n" || dl == ".\n" {
					break
				}
				// Undo dot-stuffing.
				body.WriteString(strings.TrimPrefix(dl, "."))
			}
			session.data = body.String()
			s.mu.Lock()
			s.sessions = append(s.sessions, session)
			s.mu.Unlock()
			session = &fakeSession{}
			say("250 2.0.0 queued")

		case verb == "RSET":
			session = &fakeSession{}
			say("250 2.0.0 ok")

		case verb == "QUIT":
			say("221 2.0.0 bye")
			return

		default:
			say("500 5.5.1 unrecognized")
		}
	}
}

func (s *fakeSMTP) port() float64 {
	return float64(s.ln.Addr().(*net.TCPAddr).Port)
}

func (s *fakeSMTP) last(t *testing.T) *fakeSession {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.sessions) == 0 {
		t.Fatal("server received no message")
	}
	return s.sessions[len(s.sessions)-1]
}

func (s *fakeSMTP) connections() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.connCount
}

func smtpTestAngleAddr(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "<"); i >= 0 {
		if j := strings.Index(s[i:], ">"); j >= 0 {
			return s[i+1 : i+j]
		}
	}
	return s
}

// sendMail invokes the builtin with a base option set the test can extend.
func sendMail(t *testing.T, s *fakeSMTP, opts map[string]any) (map[string]any, error) {
	t.Helper()
	base := map[string]any{
		"host": "127.0.0.1",
		"port": s.port(),
		"tls":  "none",
		"from": "app@example.com",
		"to":   "you@example.com",
	}
	for k, v := range opts {
		base[k] = v
	}
	out, err := builtinSendMail(nil, map[string]any{"0": base})
	if err != nil {
		return nil, err
	}
	result, ok := out.(map[string]any)
	if !ok {
		t.Fatalf("send_mail returned %T, want map", out)
	}
	return result, nil
}

func mustSendMail(t *testing.T, s *fakeSMTP, opts map[string]any) map[string]any {
	t.Helper()
	result, err := sendMail(t, s, opts)
	if err != nil {
		t.Fatalf("send_mail failed: %v", err)
	}
	return result
}

func TestSendMailPlainTextSinglePart(t *testing.T) {
	s := newFakeSMTP(t, "")

	result := mustSendMail(t, s, map[string]any{
		"subject": "Password reset",
		"text":    "Click the link.",
	})

	if result["ok"] != true {
		t.Errorf("ok = %v, want true", result["ok"])
	}
	id, _ := result["message_id"].(string)
	if !strings.HasPrefix(id, "<") || !strings.HasSuffix(id, "@example.com>") {
		t.Errorf("message_id = %q, want <hex@example.com>", id)
	}

	sess := s.last(t)
	if sess.from != "app@example.com" {
		t.Errorf("envelope from = %q", sess.from)
	}
	if len(sess.rcpts) != 1 || sess.rcpts[0] != "you@example.com" {
		t.Errorf("envelope rcpts = %v", sess.rcpts)
	}

	// A lone text body must not be wrapped in a pointless multipart shell.
	if strings.Contains(sess.data, "multipart") {
		t.Errorf("single-part message became multipart:\n%s", sess.data)
	}
	for _, want := range []string{
		"From: app@example.com",
		"To: you@example.com",
		"Subject: Password reset",
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=utf-8",
		"Content-Transfer-Encoding: quoted-printable",
		"Click the link.",
	} {
		if !strings.Contains(sess.data, want) {
			t.Errorf("message missing %q:\n%s", want, sess.data)
		}
	}
}

func TestSendMailMultipartAlternative(t *testing.T) {
	s := newFakeSMTP(t, "")

	mustSendMail(t, s, map[string]any{
		"subject": "Welcome",
		"text":    "Plain fallback",
		"html":    `<p>Hi <a href="https://example.com">click</a></p>`,
	})

	data := s.last(t).data
	if !strings.Contains(data, "Content-Type: multipart/alternative; boundary=") {
		t.Fatalf("expected multipart/alternative:\n%s", data)
	}

	plainAt := strings.Index(data, "text/plain")
	htmlAt := strings.Index(data, "text/html")
	if plainAt < 0 || htmlAt < 0 {
		t.Fatalf("expected both parts:\n%s", data)
	}
	// RFC 2046: least-rich alternative first, or clients show the wrong one.
	if plainAt > htmlAt {
		t.Errorf("text/html precedes text/plain in multipart/alternative:\n%s", data)
	}
	if !strings.Contains(data, "Plain fallback") {
		t.Errorf("missing text body:\n%s", data)
	}
	// The href survives quoted-printable encoding, though '=' is escaped as =3D.
	if !strings.Contains(data, "example.com") {
		t.Errorf("missing html body:\n%s", data)
	}
}

func TestSendMailHTMLOnlyIsSinglePart(t *testing.T) {
	s := newFakeSMTP(t, "")

	mustSendMail(t, s, map[string]any{"html": "<p>Hi</p>"})

	data := s.last(t).data
	if strings.Contains(data, "multipart") {
		t.Errorf("html-only message became multipart:\n%s", data)
	}
	if !strings.Contains(data, "Content-Type: text/html; charset=utf-8") {
		t.Errorf("expected text/html content type:\n%s", data)
	}
}

func TestSendMailBccIsEnvelopeOnly(t *testing.T) {
	s := newFakeSMTP(t, "")

	mustSendMail(t, s, map[string]any{
		"cc":   "cc@example.com",
		"bcc":  "secret@example.com",
		"text": "hi",
	})

	sess := s.last(t)
	want := []string{"you@example.com", "cc@example.com", "secret@example.com"}
	if len(sess.rcpts) != len(want) {
		t.Fatalf("rcpts = %v, want %v", sess.rcpts, want)
	}
	for i, addr := range want {
		if sess.rcpts[i] != addr {
			t.Errorf("rcpt[%d] = %q, want %q", i, sess.rcpts[i], addr)
		}
	}

	if !strings.Contains(sess.data, "Cc: cc@example.com") {
		t.Errorf("Cc header missing:\n%s", sess.data)
	}
	// The whole point of bcc: it reaches the recipient without appearing in headers.
	if strings.Contains(sess.data, "secret@example.com") {
		t.Errorf("bcc address leaked into headers:\n%s", sess.data)
	}
}

func TestSendMailAddressArrays(t *testing.T) {
	s := newFakeSMTP(t, "")

	// Duso arrays reach a builtin as *[]script.Value.
	recipients := []script.Value{
		script.NewString("a@example.com"),
		script.NewString("Bob <b@example.com>"),
	}
	mustSendMail(t, s, map[string]any{
		"to":   &recipients,
		"text": "hi",
	})

	sess := s.last(t)
	if len(sess.rcpts) != 2 || sess.rcpts[0] != "a@example.com" || sess.rcpts[1] != "b@example.com" {
		t.Errorf("rcpts = %v, want bare addresses for both", sess.rcpts)
	}
	// The display name belongs in the header even though the envelope strips it.
	if !strings.Contains(sess.data, `To: a@example.com, "Bob" <b@example.com>`) {
		t.Errorf("To header wrong:\n%s", sess.data)
	}
}

func TestSendMailAuthPlain(t *testing.T) {
	s := newFakeSMTP(t, "PLAIN LOGIN")

	mustSendMail(t, s, map[string]any{
		"user": "apikey",
		"pass": "secret",
		"text": "hi",
	})

	sess := s.last(t)
	if !sess.authed || sess.mech != "PLAIN" {
		t.Errorf("mech = %q authed = %v, want PLAIN true", sess.mech, sess.authed)
	}
}

func TestSendMailAuthLoginFallback(t *testing.T) {
	// LOGIN is not in net/smtp; some Exchange-flavored relays offer only this.
	s := newFakeSMTP(t, "LOGIN")

	mustSendMail(t, s, map[string]any{
		"user": "apikey",
		"pass": "secret",
		"text": "hi",
	})

	sess := s.last(t)
	if !sess.authed || sess.mech != "LOGIN" {
		t.Errorf("mech = %q authed = %v, want LOGIN true", sess.mech, sess.authed)
	}
}

func TestSendMailDialsPerSend(t *testing.T) {
	s := newFakeSMTP(t, "")

	for i := range 3 {
		mustSendMail(t, s, map[string]any{"text": fmt.Sprintf("message %d", i)})
	}

	s.mu.Lock()
	got := len(s.sessions)
	s.mu.Unlock()
	if got != 3 {
		t.Fatalf("server received %d messages, want 3", got)
	}
	// Connections are deliberately not pooled: one dial per send keeps
	// concurrent senders from serializing on a shared SMTP session.
	if conns := s.connections(); conns != 3 {
		t.Errorf("opened %d connections for 3 sends, want 3", conns)
	}
}

func TestSendMailConcurrentSends(t *testing.T) {
	s := newFakeSMTP(t, "")

	var wg sync.WaitGroup
	errs := make([]error, 4)
	for i := range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errs[i] = sendMail(t, s, map[string]any{"text": fmt.Sprintf("concurrent %d", i)})
		}()
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Errorf("concurrent send %d failed: %v", i, err)
		}
	}
	s.mu.Lock()
	got := len(s.sessions)
	s.mu.Unlock()
	if got != 4 {
		t.Errorf("server received %d messages, want 4", got)
	}
}

func TestSendMailSubjectEncoding(t *testing.T) {
	s := newFakeSMTP(t, "")

	mustSendMail(t, s, map[string]any{
		"subject": "Café — résumé",
		"text":    "hi",
	})

	data := s.last(t).data
	// Non-ASCII headers must be RFC 2047 encoded-words, never raw UTF-8.
	if strings.Contains(data, "Café") {
		t.Errorf("subject sent as raw UTF-8:\n%s", data)
	}
	if !strings.Contains(data, "=?utf-8?q?") {
		t.Errorf("subject not RFC 2047 encoded:\n%s", data)
	}
}

// headerLines returns the header block split into lines, stopping at the body.
func headerLines(data string) []string {
	head, _, _ := strings.Cut(strings.ReplaceAll(data, "\r\n", "\n"), "\n\n")
	return strings.Split(head, "\n")
}

// hasHeader reports whether a real header line with this name exists — as
// opposed to the name merely appearing inside some other header's value.
func hasHeader(data, name string) bool {
	for _, line := range headerLines(data) {
		if strings.HasPrefix(strings.ToLower(line), strings.ToLower(name)+":") {
			return true
		}
	}
	return false
}

func TestSendMailHeaderInjectionBlocked(t *testing.T) {
	s := newFakeSMTP(t, "")

	mustSendMail(t, s, map[string]any{
		"headers": map[string]any{
			"Reply-To":     "reply@example.com",
			"X-Evil":       "ok\r\nBcc: attacker@example.com",
			"X-Bad\r\nBcc": "attacker@example.com",
			"Content-Type": "text/evil",
			"Bcc":          "attacker@example.com",
		},
		"text": "hi",
	})

	sess := s.last(t)
	if !hasHeader(sess.data, "Reply-To") {
		t.Errorf("Reply-To missing:\n%s", sess.data)
	}
	// The injected value must survive only as one folded header value.
	if hasHeader(sess.data, "Bcc") {
		t.Errorf("injection produced a real Bcc header:\n%s", sess.data)
	}
	// Headers the builtin owns cannot be overridden by the caller.
	if !strings.Contains(sess.data, "Content-Type: text/plain; charset=utf-8") ||
		strings.Contains(sess.data, "text/evil") {
		t.Errorf("caller overrode Content-Type:\n%s", sess.data)
	}
	if len(sess.rcpts) != 1 {
		t.Errorf("injected header changed the envelope: %v", sess.rcpts)
	}
}

func TestSendMailRejectsControlCharsInAddress(t *testing.T) {
	s := newFakeSMTP(t, "")

	// No spaces, so a naive bare-address check would wave this through and let
	// it forge both an SMTP command and a header.
	_, err := sendMail(t, s, map[string]any{
		"to":   "a@example.com\r\nBcc:attacker@example.com",
		"text": "hi",
	})
	if err == nil {
		t.Fatal("expected an error for an address containing CRLF")
	}
	if !strings.Contains(err.Error(), "control characters") {
		t.Errorf("error = %v, want mention of control characters", err)
	}
}

func TestSendMailValidation(t *testing.T) {
	s := newFakeSMTP(t, "")

	cases := []struct {
		name string
		opts map[string]any
		want string
	}{
		{"no body", map[string]any{}, "text or html"},
		{"no to", map[string]any{"to": "", "text": "hi"}, "to address"},
		{"no host", map[string]any{"host": "", "text": "hi"}, "host"},
		{"no from", map[string]any{"from": "", "text": "hi"}, "from address"},
		{"bad recipient", map[string]any{"to": "not an address", "text": "hi"}, "recipient"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := sendMail(t, s, tc.opts)
			if err == nil {
				t.Fatalf("expected an error mentioning %q", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %v, want mention of %q", err, tc.want)
			}
		})
	}
}
