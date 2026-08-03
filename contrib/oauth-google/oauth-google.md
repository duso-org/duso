# oauth-google

Google sign-in for Duso. Built on the [oauth](/contrib/oauth/oauth.md) factory.

Google is a full OpenID Connect provider, so identity comes from a signed
`id_token` that the oauth module verifies against Google's JWKS — signature,
issuer, audience, expiry, and nonce. There is no extra API call to fetch the
user, and no vendor-specific code in this module beyond configuration.

## Setup

Create an OAuth 2.0 Client ID in the
[Google Cloud Console](https://console.cloud.google.com/apis/credentials) under
**APIs & Services → Credentials**, add your callback to **Authorized redirect
URIs**, then:

```bash
export GOOGLE_CLIENT_ID=xxxxxxxx.apps.googleusercontent.com
export GOOGLE_CLIENT_SECRET=GOCSPX-xxxxxxxxxxxx
export OAUTH_REDIRECT_URI=https://myapp.com/auth/google/callback
```

## Quick Start

Resolve Google's endpoints once at startup. Without this every `client()` call
fetches the discovery document — an HTTP round-trip to Google on every request.

```duso
// config.du — loaded once at startup
oauth = require("oauth")
google = require("oauth-google")
config = datastore("myapp")

config.set("OAUTH_GOOGLE_CLIENT_ID", env("GOOGLE_CLIENT_ID"))
config.set("OAUTH_GOOGLE_CLIENT_SECRET", env("GOOGLE_CLIENT_SECRET"))
config.set("OAUTH_GOOGLE_REDIRECT_URI", env("OAUTH_REDIRECT_URI"))

endpoints = oauth.discover(google.discovery_url)
config.set("OAUTH_GOOGLE_AUTHORIZE_URL", endpoints.authorize_url)
config.set("OAUTH_GOOGLE_TOKEN_URL", endpoints.token_url)
config.set("OAUTH_GOOGLE_JWKS_URI", endpoints.jwks_uri)
config.set("OAUTH_GOOGLE_ISSUER", endpoints.issuer)

for required in ["OAUTH_GOOGLE_CLIENT_ID", "OAUTH_GOOGLE_CLIENT_SECRET", "OAUTH_GOOGLE_REDIRECT_URI"] do
  if not config.get(required) then
    throw("config: {{required}} is not set")
  end
end
```

```duso
// server.du
include("config.du")

server = http_server({port = 8080})
server.route("GET", "/auth/google", "login.du")
server.route("GET", "/auth/google/callback", "callback.du")
server.start()
```

Each handler is its own isolated script instance, so both build the client the
same way, reading the settings back out of the datastore:

```duso
google = require("oauth-google")
config = datastore("myapp")

client = google.client({
  client_id     = config.get("OAUTH_GOOGLE_CLIENT_ID"),
  client_secret = config.get("OAUTH_GOOGLE_CLIENT_SECRET"),
  redirect_uri  = config.get("OAUTH_GOOGLE_REDIRECT_URI"),
  authorize_url = config.get("OAUTH_GOOGLE_AUTHORIZE_URL"),
  token_url     = config.get("OAUTH_GOOGLE_TOKEN_URL"),
  jwks_uri      = config.get("OAUTH_GOOGLE_JWKS_URI"),
  issuer        = config.get("OAUTH_GOOGLE_ISSUER")
})
```

```duso
// login.du
ctx = context()
ctx.response().redirect(client.authorize_url())
```

```duso
// callback.du
ctx = context()
result = client.callback(ctx)

sid = uuid()
datastore("sessions").set(sid, {
  id = result.user.id,
  email = result.user.email,
  name = result.user.name
})
datastore("sessions").expire(sid, 86400)

ctx.response().redirect("/", 302, {
  "Set-Cookie" = "sid={{sid}}; Path=/; HttpOnly; Secure; SameSite=Lax; Max-Age=86400"
})
```

**`SameSite=Lax` matters.** Google's callback is a cross-site top-level
navigation, so a `SameSite=Strict` cookie is not sent and the login silently
appears to fail.

## client(config)

| Key | Default | Description |
|---|---|---|
| `client_id` | `env("GOOGLE_CLIENT_ID")` | OAuth client ID |
| `client_secret` | `env("GOOGLE_CLIENT_SECRET")` | OAuth client secret |
| `redirect_uri` | — | Required; must match an Authorized redirect URI |
| `scope` | `openid email profile` | Space-separated scopes |
| `authorize_params` | `{access_type = "offline", prompt = "consent"}` | Replaces the defaults entirely |
| `authorize_url`, `token_url`, `jwks_uri`, `issuer` | — | Pre-resolved endpoints; supplying these skips discovery |
| `discovery` | Google's discovery URL | Override the discovery document |
| `store` | `datastore("oauth")` | Where pending logins are held |

Also exports `discovery_url` for resolving endpoints at startup.

Returns a client with `authorize_url()`, `callback(ctx)`, and `refresh(token)` —
the same surface every oauth vendor module exposes.

## The user object

```duso
result.user.id        // Google's stable subject identifier
result.user.email
result.user.name
result.user.picture
result.user.profile   // all verified id_token claims
```

These come from a **verified** `id_token`, not from an unauthenticated API call.

## Refresh tokens

Google issues a refresh token only when `access_type=offline` is sent **and** the
user actively consents. A user who has already approved your app is normally sent
straight through without a consent screen, and comes back with **no refresh
token** — which is why `prompt=consent` is on by default here.

If you do not need offline access, drop it:

```duso
client = google.client({
  redirect_uri = config.get("OAUTH_GOOGLE_REDIRECT_URI"),
  authorize_params = {access_type = "online"}
})
```

Note that `authorize_params` replaces the defaults rather than merging, so this
drops `prompt=consent` too. For per-login additions, use `authorize_url()`:

```duso
client.authorize_url({params = {login_hint = "user@example.com"}})
```

## Restricting to a Google Workspace domain

Google accepts an `hd` parameter as a hint, but **it is not a security control** —
verify the claim yourself after login:

```duso
result = client.callback(ctx)
if result.user.profile.hd != "mycorp.com" then
  ctx.response().error(403, "This application is restricted to mycorp.com accounts")
end
```

## See Also

- [oauth - the client factory and vendor contract](/contrib/oauth/oauth.md)
- [oauth-github - the non-OIDC counterpart](/contrib/oauth-github/oauth-github.md)
- [http_server() - cookies and response headers](/docs/reference/http_server.md)
