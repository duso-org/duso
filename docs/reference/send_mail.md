# send_mail()

Send an email through an SMTP relay. Duso sends mail; it never receives mail and never acts as a mail server. This opens an outbound connection to a relay you configure, hands over one message, and returns.

`send_mail(options)`

## Parameters

`options` (object) - a single flat object:

**Connection**

- `host` (string, required) - relay hostname
- `port` (number) - default 587
- `user` (string) - AUTH username; omit for an unauthenticated relay
- `pass` (string) - AUTH password or API key
- `tls` (string) - `"auto"` (default), `"implicit"`, `"starttls"`, or `"none"`
- `timeout` (number) - seconds for the whole send, default 30

**Message**

- `from` (string, required) - sender, plain or `"Name <a@b.com>"`
- `to` (string or array, required) - recipients
- `cc` (string or array) - copied recipients
- `bcc` (string or array) - blind recipients; delivered but never written into the headers
- `subject` (string) - encoded automatically when it contains non-ASCII
- `text` (string) - the `text/plain` body
- `html` (string) - the `text/html` body
- `headers` (object) - extra headers such as `Reply-To`

At least one of `text` or `html` is required. When both are given the message is
sent as `multipart/alternative` and the client picks; when only one is given the
message is a single part.

## Returns

Object with:

- `ok` (boolean) - true when the relay accepted the message
- `message_id` (string) - the generated `Message-ID` header

Failure to connect, authenticate, or deliver raises an error, so wrap sends in
`try`/`catch` the same way you would a `fetch()`.

## Examples

Simplest send:

```duso
send_mail({
  host = "smtp-relay.brevo.com",
  user = "myaccount",
  pass = env("SMTP_PASS"),
  from = "app@example.com",
  to = "you@example.com",
  subject = "Password reset",
  text = "Click the link: https://example.com/reset"
})
```

Keeping credentials out of every call site. Store the connection once, then use
the object as a constructor to compose each message from it:

```duso
mail = datastore("config").get("relay")

send_mail(mail(
  to = "you@example.com",
  subject = "Welcome",
  text = "Thanks for signing up."
))
```

Text and HTML together, which is what most transactional mail should send:

```duso
send_mail(mail(
  to = "you@example.com",
  subject = "Your receipt",
  text = "Thanks! View your receipt: https://example.com/receipt/123",
  html = """
    <p>Thanks!</p>
    <p><a href="https://example.com/receipt/123">View your receipt</a></p>
  """
))
```

Multiple recipients, a blind copy, and a reply address:

```duso
send_mail(mail(
  to      = ["a@example.com", "Bob <b@example.com>"],
  cc      = "team@example.com",
  bcc     = "audit@example.com",
  headers = {["Reply-To"] = "support@example.com"},
  subject = "Deploy finished",
  text    = "Build 412 is live."
))
```

Handling failure:

```duso
try
  send_mail(mail(to = user.email, subject = "Reset", text = body))
catch (err)
  print("mail failed: " + err.message)
end
```

## Connections

Each call opens its own connection, sends, and closes it. Connections are not
pooled: relays drop idle sessions within 30 to 60 seconds, so a cache would miss
for any app that sends occasionally, and a shared session would force concurrent
senders to queue behind one another. Sends from separate `spawn()` calls therefore
run genuinely in parallel.

The handshake costs a few hundred milliseconds, which is small next to the request
that triggered the mail. If you are sending in a tight loop and that cost matters,
`spawn()` the sends so they overlap.

A send inherits the cancellation context of its process, so `kill(pid)` aborts one
in flight, and `timeout` bounds the whole exchange.

## TLS

With `tls = "auto"` the mode follows the port: 465 negotiates TLS immediately,
anything else connects and upgrades with STARTTLS when the relay offers it. Set
`tls` explicitly when a relay disagrees with that convention. `"none"` sends in the
clear and should only be used for a relay on localhost.

`AUTH PLAIN`, `LOGIN`, and `CRAM-MD5` are supported; the mechanism is chosen from
what the relay advertises.

## Notes on deliverability

Sending mail is easy; getting it delivered is the hard part, and it is largely
outside duso's control.

- Send through a relay you have configured with SPF, DKIM, and DMARC for your
  sending domain. Mail sent straight from an application server without those
  records is usually filtered as spam.
- Provide a `text` body alongside `html`. HTML-only messages score worse with
  spam filters.
- Images must be hosted and linked with a normal `https` URL. Data URIs
  (`<img src="data:image/png;base64,...">`) are stripped by Gmail and unsupported
  by Outlook, so a logo embedded that way is invisible to most recipients.
  Remote images are shown by default in Gmail and blocked until the reader clicks
  in Outlook, which is the normal tradeoff for transactional mail.

## See also

- [`fetch()`](/docs/reference/fetch.md) - for relays that expose an HTTP API instead
- [`http_server()`](/docs/reference/http_server.md) - to receive bounce and delivery webhooks
- [`env()`](/docs/reference/env.md) - to keep credentials out of source
