# DigitalOcean API Client for Duso

Droplets and the resources you need to create one: sizes, regions, images, SSH
keys, DNS records, snapshots. Enough to provision a server, wait for it to boot,
point a hostname at it, and later upgrade or destroy it.

## Installation

```duso
ocean = require("digitalocean")
```

## Quick start

```duso
ocean = require("digitalocean")

droplet = ocean.droplets.create({
  name = "web-1",
  region = "sfo3",
  size = "s-1vcpu-512mb-10gb",
  image = "ubuntu-24-04-x64",
  ssh_keys = [52468621]
}, {wait = true})

print("{{droplet.name}} is up at {{ocean.public_ip(droplet)}}")
```

## Authentication

Set a Personal Access Token, created in the DigitalOcean control panel under
**API → Tokens**:

```bash
export DIGITALOCEAN_TOKEN="dop_v1_..."
duso provision.du
```

`DIGITALOCEAN_ACCESS_TOKEN` is also read, so a machine already set up for `doctl`
or Terraform needs nothing new. To use a different token for one call, pass it:

```duso
droplet = ocean.droplets.get(3164444, {token = "dop_v1_..."})
```

> Prefer a scoped token. A blanket read/write token can destroy every droplet on
> the account and rewrite its DNS. Provisioning needs `droplet:create`,
> `droplet:read`, `droplet:update`, `droplet:delete`, `ssh_key:read`,
> `image:read` and, if it writes DNS, `domain:read` and `domain:update`. Add
> `tag:create` only if you pass `tags` when creating — a token without it makes
> the whole create fail, not just the tagging.

## Calling convention

Every function takes API-shaped data first and an optional `opts` object last.

```duso
ocean.droplets.create(spec, opts)
ocean.droplets.resize(id, "s-2vcpu-2gb", opts)
ocean.domains.records.create("example.com", record, opts)
```

`opts` accepts:

| key | default | meaning |
|---|---|---|
| `token` | `DIGITALOCEAN_TOKEN` | credential for this call |
| `timeout` | `30` | HTTP timeout in seconds |
| `retries` | `3` | attempts on 429, 5xx, and dropped connections |
| `wait` | `false` | block until the action or droplet settles |
| `interval` | `3` | seconds between polls while waiting |
| `disk` | `false` | on `resize` only — grow the disk permanently |

Request bodies pass through to the API unchanged rather than being mapped onto
invented parameter names, so the
[DigitalOcean reference](https://docs.digitalocean.com/reference/api/reference/)
is the reference for this module too, and a field they add works the day they
add it.

## Droplets

```duso
droplets = ocean.droplets.list()
droplets = ocean.droplets.list({tag_name = "web"})

droplet = ocean.droplets.get(3164444)

droplet = ocean.droplets.create({
  name = "web-1",
  region = "sfo3",
  size = "s-1vcpu-512mb-10gb",
  image = "ubuntu-24-04-x64",
  ssh_keys = [52468621],
  user_data = load("cloud-init.yaml"),
  monitoring = true,
  tags = ["web"]
})

ocean.droplets.delete(3164444)
```

Lists are fully paginated — you get every droplet, not the first twenty.

`create` returns as soon as DigitalOcean accepts the request, which is before
the machine exists: the droplet comes back with status `new` and no address.

### Addresses

```duso
ocean.public_ip(droplet)     // "143.198.63.21", or nil before boot
ocean.private_ip(droplet)    // "10.124.0.4", or nil
```

### Waiting

Three different things settle at three different times, so there are three
waiters.

```duso
droplet = ocean.droplets.wait_ready(3164444)
```

`wait_ready` blocks until the droplet is `active` **and** has a public address.
Both, because networking lands a moment after the status flips, and code that
writes a DNS record on status alone writes a `nil`. Typically 35–45 seconds.

```duso
action = ocean.droplets.power_off(3164444, {wait = true})
```

`{wait = true}` blocks until the *action* reports `completed`.

```duso
droplet = ocean.droplets.wait_status(3164444, "off")
```

`wait_status` blocks until the *droplet* agrees. These are not the same event:
a completed `power_off` is routinely followed by a droplet that still reads
`active` for several more seconds. Anything that branches on status after an
action — including `resize`, which is refused while the droplet reads active —
needs `wait_status`, not `wait`.

```duso
ocean.droplets.power_off(id, {wait = true})
ocean.droplets.wait_status(id, "off")
ocean.droplets.resize(id, "s-2vcpu-2gb", {wait = true})
ocean.droplets.power_on(id, {wait = true})
ocean.droplets.wait_status(id, "active")
```

All three waiters throw `{id = "timeout"}` rather than returning something
half-finished. Default ceiling is 300 seconds; override with `{timeout = 600}`.

> Don't wait inside an HTTP handler. A create-and-wait holds a request for a
> minute. Provision from a `spawn()` and report through a `datastore()`.

### Actions

```duso
ocean.droplets.resize(id, "s-2vcpu-2gb")
ocean.droplets.reboot(id)
ocean.droplets.power_on(id)
ocean.droplets.power_off(id)
ocean.droplets.power_cycle(id)
ocean.droplets.shutdown(id)
ocean.droplets.rebuild(id, "ubuntu-24-04-x64")
ocean.droplets.rename(id, "web-2")
ocean.droplets.snapshot(id, "before upgrade")
ocean.droplets.enable_backups(id)
ocean.droplets.enable_ipv6(id)
```

Each returns the action object. Add `{wait = true}` to block until it settles.

> `resize` leaves `disk` false unless you ask for it. A resize that grows the
> disk is **permanent** — the droplet can never be resized down afterwards. CPU
> and memory alone are reversible.

Any action type, including ones newer than this module:

```duso
action = ocean.droplets.action(id, {type = "enable_backups"}, {wait = true})
```

And to poll one you already have:

```duso
action = ocean.actions.get(36804636)
action = ocean.actions.wait(36804636)
```

`actions.wait` throws `{id = "action_errored"}` if DigitalOcean reports the
action failed.

## Sizes, regions, images

```duso
sizes = ocean.sizes.list()
regions = ocean.regions.list()

images = ocean.images.list({type = "distribution"})
images = ocean.images.list({private = true})
image = ocean.images.get("ubuntu-24-04-x64")
```

Pick the cheapest plan that actually exists in your region rather than hardcoding
a slug:

```duso
ocean = require("digitalocean")

available = filter(ocean.sizes.list(), function(s)
  return s.available and contains(join(s.regions, " "), "sfo3")
end)

cheapest = sort(available, function(a, b)
  return a.price_monthly < b.price_monthly
end)[0]

print("{{cheapest.slug}} at ${{cheapest.price_monthly}}/mo")
```

## SSH keys

```duso
ssh_keys = ocean.ssh_keys.list()
key = ocean.ssh_keys.get(52468621)
key = ocean.ssh_keys.create({name = "laptop", public_key = "ssh-ed25519 AAAA..."})
ocean.ssh_keys.delete(52468621)
```

A droplet created with no `ssh_keys` emails a root password instead, so pass one.

## Snapshots

```duso
snapshots = ocean.snapshots.list()
snapshots = ocean.snapshots.list({resource_type = "droplet"})
snapshot = ocean.snapshots.get("119192817")
ocean.snapshots.delete("119192817")
```

Taking one is a droplet action:

```duso
ocean.droplets.snapshot(3164444, "nightly", {wait = true})
```

## Tags

```duso
tags = ocean.tags.list()
tag = ocean.tags.get("org:7f3a91")
tag = ocean.tags.create("org:7f3a91")
ocean.tags.delete("org:7f3a91")

ocean.tags.attach("org:7f3a91", 3164444)
ocean.tags.attach("org:7f3a91", [3164444, 3164445])
ocean.tags.detach("org:7f3a91", 3164444)
```

Names may contain letters, numbers, colons, dashes and underscores, up to 255
characters, and are **case stable** — DigitalOcean keeps the capitalisation you
first used and expects an exact match afterwards. Lowercase anything derived
from user input.

Attach/detach default to droplets. For other resources:

```duso
ocean.tags.attach("backups", "3d80cb72-342b-4aaa-b92e-4e4abb24a933", {resource_type = "volume"})
```

Tagging at create time is simpler when you know the tags up front, and creates
any that don't exist yet:

```duso
droplet = ocean.droplets.create({
  name = "hjwfn7xs",
  region = "sfo3",
  size = "s-1vcpu-1gb",
  image = "ubuntu-24-04-x64",
  tags = ["arland", "org:7f3a91"]
})
```

> That implicit creation is why passing `tags` needs `tag:create` on the token.
> Without it the **whole create fails**, not just the tagging.

Filtering is one tag at a time — the API has no AND:

```duso
mine = ocean.droplets.list({tag_name = "org:7f3a91"})
```

## Domains and DNS

```duso
domains = ocean.domains.list()
domain = ocean.domains.get("example.com")
domain = ocean.domains.create({name = "example.com"})
ocean.domains.delete("example.com")
```

Records take the API's own shape:

```duso
records = ocean.domains.records.list("example.com")
records = ocean.domains.records.list("example.com", {type = "A"})

record = ocean.domains.records.create("example.com", {
  type = "A",
  name = "www",
  data = "143.198.63.21",
  ttl = 1800
})

ocean.domains.records.update("example.com", 28448433, {type = "A", data = "10.0.0.1"})
ocean.domains.records.delete("example.com", 28448433)
```

`upsert` points a name at something whether or not it already points somewhere,
matching on type and name:

```duso
ocean.domains.records.upsert("example.com", {
  type = "A",
  name = "www",
  data = "143.198.63.21",
  ttl = 1800
})
```

Use it in provisioning. Runs get retried — a half-finished attempt, a customer
clicking twice — and plain `create` would leave two A records for one host,
round-robining traffic to a machine that may no longer exist.

## Account

```duso
acct = ocean.account.get()
print("{{acct.email}} — {{acct.droplet_limit}} droplet limit")
```

Requires `account:read`, which a narrowly scoped provisioning token won't have.

## Errors

Failures throw an object, because a caller has to tell "this droplet is gone"
from "slow down" from "that size slug doesn't exist", and those are three
different recoveries.

```duso
try
  droplet = ocean.droplets.get(999)
catch (e)
  print("{{e.status}} {{e.id}}: {{e.message}}")
end
```

| field | value |
|---|---|
| `status` | HTTP status, or `0` for failures that never reached the API |
| `id` | DigitalOcean's error id (`not_found`, `unprocessable_entity`), or one of this module's: `no_token`, `unreachable`, `timeout`, `invalid`, `action_errored` |
| `message` | human-readable, safe to show a user |
| `request_id` | DigitalOcean's id for the request, for support tickets |

Branching on it:

```duso
ocean = require("digitalocean")

function ensure_gone(id)
  try
    ocean.droplets.delete(id)
  catch (e)
    if e.status == 404 then
      return "already gone"
    end
    throw(e)
  end
  return "deleted"
end

print(ensure_gone(3164444))
```

## Rate limits and retries

DigitalOcean allows 5,000 requests an hour. A 429, a 5xx, or a dropped
connection is retried up to `retries` times with exponential backoff. The
`ratelimit-reset` header only ever *shortens* the wait — it can be most of an
hour away, which is longer than any caller wants to block — and backoff is
capped at 30 seconds regardless.

4xx responses other than 429 are not retried; they will not succeed on a second
attempt and retrying them only spends the budget.

```duso
droplet = ocean.droplets.get(3164444, {retries = 0})
```

## Pagination

Every `list` follows `links.pages.next` and returns one flat array, requesting
the API's maximum 200 per page. There is no page parameter to manage.

## Escape hatch

Anything this module doesn't wrap:

```duso
result = ocean.request("GET", "/vpcs", nil, {per_page = 20})
```

`request(method, path, body, query, opts)` handles auth, retries, JSON and error
throwing. It returns the parsed body, unwrapped from nothing — you get
`result.vpcs`, not `result`.

## Example: provision a server and point a hostname at it

```duso
ctx = context()
ocean = require("digitalocean")
store = datastore("provisioning")

droplet = ocean.droplets.create({
  name = ctx.host,
  region = "sfo3",
  size = "s-1vcpu-512mb-10gb",
  image = "ubuntu-24-04-x64",
  ssh_keys = [ctx.ssh_key_id],
  user_data = load("cloud-init.yaml"),
  monitoring = true
}, {wait = true})

ip = ocean.public_ip(droplet)

ocean.domains.records.upsert("example.com", {
  type = "A",
  name = ctx.host,
  data = ip,
  ttl = 300
})

store.set("host_" + ctx.host, {id = droplet.id, ip = ip})
```

Run it with `spawn("provision.du", {host = "web-1", ssh_key_id = 52468621})` so
the minute it spends waiting isn't a minute a request is blocked.

## Example: upgrade in place

```duso
ocean = require("digitalocean")

function upgrade(id, size)
  ocean.droplets.power_off(id, {wait = true})
  ocean.droplets.wait_status(id, "off")

  ocean.droplets.resize(id, size, {wait = true})

  ocean.droplets.power_on(id, {wait = true})
  return ocean.droplets.wait_status(id, "active")
end

droplet = upgrade(3164444, "s-1vcpu-1gb")
print("now {{droplet.size_slug}} at {{ocean.public_ip(droplet)}}")
```

The address survives a resize. Roughly 70 seconds end to end.

## Limitations

- No firewalls, volumes, load balancers, reserved IPs, VPCs, Kubernetes or
  databases. Reach them through `ocean.request()`.
- Multi-droplet create (`names` rather than `name`) is not wrapped; the response
  shape differs and the waiters assume one droplet.
- Deleting a droplet has no undo and no trash.
