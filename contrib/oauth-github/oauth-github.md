# oauth-github

GitHub OAuth 2.0 login for Duso. Built on the [oauth](/contrib/oauth/oauth.md)
factory — the flow, PKCE, and state handling are shared; this module supplies
GitHub's endpoints and its user lookup.

## Setup

Register an OAuth app at **Settings → Developer settings → OAuth Apps**, set the
callback URL to your `/auth/github/callback` route, then:

```bash
export GITHUB_CLIENT_ID=Iv1.xxxxxxxxxxxx
export GITHUB_CLIENT_SECRET=xxxxxxxxxxxxxxxxxxxxxxxx
export OAUTH_REDIRECT_URI=https://myapp.com/auth/github/callback
```

## Quick Start

Load configuration once at startup into a datastore, then read it from the
handlers. Each handler is an isolated script instance with no shared scope, and
each `require()` returns a fresh module instance, so a datastore is where
per-application settings belong.

```duso
// config.du — loaded once at startup
config = datastore("myapp")

config.set("OAUTH_GITHUB_CLIENT_ID", env("GITHUB_CLIENT_ID"))
config.set("OAUTH_GITHUB_CLIENT_SECRET", env("GITHUB_CLIENT_SECRET"))
config.set("OAUTH_GITHUB_REDIRECT_URI", env("OAUTH_REDIRECT_URI"))

// Fail the boot, not a user's first login
for required in ["OAUTH_GITHUB_CLIENT_ID", "OAUTH_GITHUB_CLIENT_SECRET", "OAUTH_GITHUB_REDIRECT_URI"] do
  if not config.get(required) then
    throw("config: {{required}} is not set")
  end
end
```

```duso
// server.du
include("config.du")

server = http_server({port = 8080})
server.route("GET", "/auth/github", "login.du")
server.route("GET", "/auth/github/callback", "callback.du")
server.start()
```

```duso
// login.du
gh = require("oauth-github")
config = datastore("myapp")
ctx = context()

client = gh.client({
  client_id     = config.get("OAUTH_GITHUB_CLIENT_ID"),
  client_secret = config.get("OAUTH_GITHUB_CLIENT_SECRET"),
  redirect_uri  = config.get("OAUTH_GITHUB_REDIRECT_URI")
})

ctx.response().redirect(client.authorize_url())
```

```duso
// callback.du
gh = require("oauth-github")
config = datastore("myapp")
ctx = context()

client = gh.client({
  client_id     = config.get("OAUTH_GITHUB_CLIENT_ID"),
  client_secret = config.get("OAUTH_GITHUB_CLIENT_SECRET"),
  redirect_uri  = config.get("OAUTH_GITHUB_REDIRECT_URI")
})

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

Then read the session in any handler:

```duso
ctx = context()
sid = ctx.request().cookies.sid
session = sid ? datastore("sessions").get(sid) : nil

if session == nil then
  ctx.response().redirect("/auth/github")
end
```

**`SameSite=Lax` matters.** GitHub's callback is a cross-site top-level navigation,
so a `SameSite=Strict` cookie is not sent and the login silently appears to fail.

## client(config)

| Key | Default | Description |
|---|---|---|
| `client_id` | `env("GITHUB_CLIENT_ID")` | OAuth app client ID |
| `client_secret` | `env("GITHUB_CLIENT_SECRET")` | OAuth app client secret |
| `redirect_uri` | — | Required; must match the app's registered callback |
| `scope` | `read:user user:email` | Space-separated scopes |
| `base_url` | `https://github.com` | Override for GitHub Enterprise Server |
| `api_url` | `https://api.github.com` | Override for GitHub Enterprise Server |
| `store` | `datastore("oauth")` | Where pending logins are held |

Returns a client with `authorize_url()`, `callback(ctx)`, and `refresh(token)` —
the same surface every oauth vendor module exposes. See
[oauth](/contrib/oauth/oauth.md) for the full contract.

## The user object

```duso
result.user.id        // "583231" — always a string
result.user.email     // primary verified address, or nil
result.user.name      // display name, falling back to the login handle
result.user.picture   // avatar URL
result.user.profile   // GitHub's full /user response
```

**Email may be `nil`.** GitHub omits it from `/user` when the user keeps their
address private. This module then calls `/user/emails` and takes the primary
verified address, which needs the `user:email` scope. If that scope was not
granted, `email` is `nil` and the rest of the identity is still usable — so treat
email as optional rather than assuming it is present.

## GitHub Enterprise Server

```duso
client = gh.client({
  redirect_uri = env("OAUTH_REDIRECT_URI"),
  base_url = "https://github.mycorp.com",
  api_url = "https://github.mycorp.com/api/v3"
})
```

## Differences from an OIDC provider

GitHub is plain OAuth 2.0, not OpenID Connect:

- **No `id_token`.** Identity comes from an API call, not a signed token, so there
  is no signature, issuer, audience, or nonce to verify. The `state` and PKCE
  checks still apply in full.
- **Form-encoded token responses.** GitHub's token endpoint answers
  `application/x-www-form-urlencoded` unless asked otherwise; this module sends
  `Accept: application/json`, and the oauth base module reads either.
- **No refresh tokens** for standard OAuth apps. `refresh()` exists on the client
  but GitHub only issues refresh tokens for apps with token expiration enabled.

## See Also

- [oauth - the client factory and vendor contract](/contrib/oauth/oauth.md)
- [http_server() - cookies and response headers](/docs/reference/http_server.md)
- [datastore() - session storage](/docs/reference/datastore.md)
