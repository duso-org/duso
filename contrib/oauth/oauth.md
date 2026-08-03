# oauth

OAuth 2.0 client factory — Authorization Code flow with PKCE, for any OAuth 2.0 or
OIDC provider.

This module is a **factory, not a provider**. It contains the protocol and knows no
vendor names. Vendor modules like [oauth-github](/contrib/oauth-github/oauth-github.md)
call `create_client()` with their endpoints and expose a `client()` function.

Most applications use a vendor module rather than this one directly:

```duso
gh = require("oauth-github")
client = gh.client({redirect_uri = "https://myapp.com/auth/github/callback"})
```

Use this module directly for a provider that has no vendor module — anything with
OIDC discovery works with no preset at all.

## Quick Start

Load configuration once at startup into a datastore, and read it from the
handlers. Each HTTP handler is its own isolated script instance with no shared
scope, and each `require()` returns a fresh module instance, so a datastore is
where per-application settings belong.

Resolving discovery at startup matters: `create_client()` fetches the discovery
document when given a `discovery` URL, so calling it per request would mean an
HTTP round-trip to the provider on **every** request.

```duso
// config.du — loaded once at startup
oauth = require("oauth")
config = datastore("myapp")

config.set("OAUTH_CLIENT_ID", env("IDP_CLIENT_ID"))
config.set("OAUTH_CLIENT_SECRET", env("IDP_CLIENT_SECRET"))
config.set("OAUTH_REDIRECT_URI", env("OAUTH_REDIRECT_URI"))

// One discovery fetch for the life of the server
endpoints = oauth.discover("https://id.example.com/.well-known/openid-configuration")
config.set("OAUTH_AUTHORIZE_URL", endpoints.authorize_url)
config.set("OAUTH_TOKEN_URL", endpoints.token_url)
config.set("OAUTH_JWKS_URI", endpoints.jwks_uri)
config.set("OAUTH_ISSUER", endpoints.issuer)

// Fail the boot, not a user's first login
for required in ["OAUTH_CLIENT_ID", "OAUTH_CLIENT_SECRET", "OAUTH_REDIRECT_URI"] do
  if not config.get(required) then
    throw("config: {{required}} is not set")
  end
end
```

```duso
// server.du
include("config.du")

server = http_server({port = 8080})
server.route("GET", "/auth/login", "login.du")
server.route("GET", "/auth/callback", "callback.du")
server.start()
```

```duso
// login.du — send the user to the provider
oauth = require("oauth")
config = datastore("myapp")
ctx = context()

client = oauth.create_client({
  client_id     = config.get("OAUTH_CLIENT_ID"),
  client_secret = config.get("OAUTH_CLIENT_SECRET"),
  redirect_uri  = config.get("OAUTH_REDIRECT_URI"),
  authorize_url = config.get("OAUTH_AUTHORIZE_URL"),
  token_url     = config.get("OAUTH_TOKEN_URL"),
  jwks_uri      = config.get("OAUTH_JWKS_URI"),
  issuer        = config.get("OAUTH_ISSUER"),
  scope         = "openid email profile"
})

ctx.response().redirect(client.authorize_url())
```

```duso
// callback.du — complete the login
oauth = require("oauth")
config = datastore("myapp")
ctx = context()

client = oauth.create_client({
  client_id     = config.get("OAUTH_CLIENT_ID"),
  client_secret = config.get("OAUTH_CLIENT_SECRET"),
  redirect_uri  = config.get("OAUTH_REDIRECT_URI"),
  authorize_url = config.get("OAUTH_AUTHORIZE_URL"),
  token_url     = config.get("OAUTH_TOKEN_URL"),
  jwks_uri      = config.get("OAUTH_JWKS_URI"),
  issuer        = config.get("OAUTH_ISSUER")
})

result = client.callback(ctx)

sid = uuid()
datastore("sessions").set(sid, {email = result.user.email})
datastore("sessions").expire(sid, 86400)

ctx.response().redirect("/", 302, {
  "Set-Cookie" = "sid={{sid}}; Path=/; HttpOnly; Secure; SameSite=Lax; Max-Age=86400"
})
```

**`SameSite=Lax` matters.** The provider's callback is a cross-site top-level
navigation, so a `SameSite=Strict` session cookie is not sent and the login
silently appears to fail.

## create_client(config)

Returns a client object. Configuration:

| Key | Required | Description |
|---|---|---|
| `client_id` | yes | OAuth client identifier |
| `redirect_uri` | yes | Must match what the provider has registered |
| `client_secret` | usually | Omitted only for public clients |
| `discovery` | — | OIDC discovery URL; fills in the endpoints below |
| `authorize_url` | — | Authorization endpoint, if not using discovery |
| `token_url` | — | Token endpoint, if not using discovery |
| `jwks_uri` | — | Signing keys; presence of this enables ID token verification |
| `issuer` | — | Expected `iss` claim |
| `scope` | — | Default scope string |
| `authorize_params` | — | Extra parameters for the authorize redirect |
| `token_headers` | — | Extra headers for the token POST |
| `userinfo` | — | `function(access_token)` returning a user, for non-OIDC providers |
| `client_secret_fn` | — | `function()` returning the secret, for providers that require a freshly signed one |
| `store` | — | Datastore for pending logins; defaults to `datastore("oauth")` |

Either `discovery` or both `authorize_url` and `token_url` must be supplied.

`discovery` costs an HTTP request every time `create_client()` runs, which is
convenient for scripts and development but wrong for a server. Use
`discover()` at startup instead and pass the resolved endpoints.

## discover(url)

Fetches an OIDC discovery document and returns
`{authorize_url, token_url, jwks_uri, issuer}` — the settings `create_client()`
needs. Call it once at startup and keep the result in a datastore. A provider that
is unreachable or misconfigured then fails the boot instead of failing a login.

```duso
endpoints = oauth.discover("https://id.example.com/.well-known/openid-configuration")
```

## Client methods

### authorize_url([options])

Returns the URL to redirect the user to. Generates the `state`, PKCE verifier, and
nonce, and records them so `callback()` can verify the return trip.

```duso
url = client.authorize_url()
url = client.authorize_url({scope = "openid email", params = {login_hint = "a@b.com"}})
```

- `scope` — overrides the client's default scope
- `params` — extra query parameters merged into the redirect

### callback(ctx)

Takes the handler's `context()`. Validates `state`, exchanges the code (sending the
PKCE verifier), and resolves the user. Throws on any failure.

Returns `{user, tokens}`:

```duso
result = client.callback(ctx)

result.user.id           // provider's stable user identifier, always a string
result.user.email
result.user.name
result.user.picture
result.user.profile      // the provider's original claims or profile response

result.tokens.access_token
result.tokens.refresh_token   // when the provider issues one
```

`profile` holds everything the provider actually sent, so nothing is lost to
normalization.

### refresh(refresh_token)

Exchanges a refresh token for a new token set.

```duso
tokens = client.refresh(stored.refresh_token)
```

## What it verifies

Every one of these is a real attack if skipped, so `callback()` throws rather than
returning a partial result:

- **State** — an unrecognized or expired `state` is refused. This is the CSRF
  defense: it proves this server started the login.
- **Single use** — the pending record is deleted before the code exchange, so a
  replayed callback finds nothing.
- **PKCE (S256)** — the verifier never leaves the server, so an intercepted
  authorization code cannot be redeemed by anyone else. `plain` is not offered.
- **ID token signature** — checked against the provider's JWKS, RS\* or ES\*.
- **Issuer and audience** — a validly signed token from another issuer, or minted
  for a different application, is rejected.
- **Expiry** and **nonce** — the nonce binds the token to this specific login.

Pending logins expire after 10 minutes and clean themselves up.

## Writing a vendor module

A vendor module is thin. It supplies endpoints and, if the provider issues no
`id_token`, a `userinfo` function. It must not reimplement the flow.

```duso
oauth_base = require("oauth")

function client(config)
  if not config then config = {} end
  return oauth_base.create_client({
    client_id = config.client_id or env("MYPROVIDER_CLIENT_ID"),
    client_secret = config.client_secret or env("MYPROVIDER_CLIENT_SECRET"),
    redirect_uri = config.redirect_uri,
    scope = config.scope or "openid email",
    discovery = "https://provider.example/.well-known/openid-configuration",
    store = config.store
  })
end

return {client = client}
```

See [oauth-google](/contrib/oauth-google/oauth-google.md) for a worked OIDC
example. A provider with no OIDC supplies explicit endpoints plus a `userinfo`
function returning `{id, email, name, picture, profile}` — see
[oauth-github](/contrib/oauth-github/oauth-github.md).

Every vendor module exposes the same `client()` surface, so providers are
interchangeable at the call site and an application can select one at runtime.

## Also exported

For vendor modules that need them rather than reimplementing them:

- `b64url(data)` / `b64url_decode(str)` — base64url per RFC 4648 section 5
- `random_token()` — 64-character token suitable for state and PKCE verifiers
- `secure_equals(a, b)` — constant-time comparison

## See Also

- [oauth-github](/contrib/oauth-github/oauth-github.md)
- [http_server() - Custom headers and cookies](/docs/reference/http_server.md)
- [format_form() / parse_form()](/docs/reference/format_form.md)
