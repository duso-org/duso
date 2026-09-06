# Duso: The First Seven Months

*State of the project, September 2026*

## Introduction

**It's been fun but also super lonely.** I can't be the only dev that has
grown tired of the direction we've been going. I hope to hear from you, even if
you just want to let me know you like some of what you see in Duso.

I wanted server-side development to feel good again, the way it did before you
had to assemble forty pieces to get to hello world. I didn't see what I wanted,
so I built it.

I made the first commit to Duso on January 22, 2026. Seven months later it's a
scripting language and a server runtime in one 12MB binary: HTTP server, ACID
datastore, WebSockets, AI clients, image processing, a debugger, a linter, an
LSP, and the documentation. It idles at about 5MB of RAM.

Thanks to [Shannan.dev](https://shannan.dev) and
[Ludonode](https://ludonode.com), who sponsor the work.

## The Idea

Somewhere along the line, starting a web server turned into a shopping trip.
Pick a runtime, a framework, a package manager, a bundler, a process manager,
something for caching, something for migrations, and then write some code,
assuming all the versions still agree with each other by the time you're done.

Duso is one binary that already made those choices. Web server, routing,
templates, an ACID datastore, SQL drivers, WebSockets, crypto, image processing,
AI clients, a debugger, a linter, an LSP, and the docs. Nothing to install,
nothing to assemble, nothing to keep in sync.

A few things fall out of that. The language stays boring on purpose, with one
obvious way to do each thing. There's no async/await, so `fetch()` just returns
the response, and when you want concurrency you `spawn()` a script or run
functions in `parallel()`. That's Go's concurrency without writing Go. Being
predictable is also why AI tends to write decent Duso on the first try, since
there's no trivia to trip over.

Dev and prod are the same binary run the same way. Route handlers are plain
scripts, checked for changes on every request, so you save the file and hit
refresh. No watcher, no bundler, no restart.

The batteries aren't a metaphor either. The datastore covers what most apps
reach for Redis to do. The docs work on a plane. Untrusted code can be
sandboxed, which matters more every month that people paste in scripts a model
wrote thirty seconds ago. When you're finished, `bundle-duso` turns your app
into its own executable and deploying is `scp` and run.

1.x is about nailing single-node capability and vertical scaling, which covers a
lot more ground than people give it credit for. 2.x will be about nodes working
in concert, and scaling horizontally.

## Major Milestones

**January: the runtime shows up.** The first month wasn't a prototype. By the
end of it Duso had an interpreter, a module system, `http_server()`,
`datastore()`, and the concurrency story: `spawn()`, `run()`, `parallel()`. It
also had a debugger and its own docs baked into the binary, because I get
annoyed looking things up. February brought the virtual filesystem, `fetch()`,
mutable arrays, and an LSP server. The VS Code extension went up the next day.

**February and March: it becomes a web server.** Static file serving meant
`duso eval 'http_server().start()'` was suddenly a usable dev server. The
Homebrew tap went up on the 19th, same day as the first tagged release. In March
I added a Tree-sitter grammar and a Zed extension, then Sublime a few days
later. OpenAI landed and brought Ollama, DeepSeek, Azure AI, and Groq along with
it. Markdown got promoted from a module to a builtin, since AI reads and writes
the stuff constantly. `parse()` and the `code` type showed up so you could hand
an AST to `run()` or `spawn()`, which is what makes letting an agent write code
actually workable. Security arrived in a batch: CORS, JWT, password hashing,
base64. Then the `binary` type, file uploads, a 33% speedup, and WebSockets.

**May: 1.0.** The datastore became ACID, with write-ahead logging, atomic
`update()`, and a real binary format on disk. `bundle-duso` let you compile your
own scripts into a standalone binary. The language got a formal spec so it could
stand on its own. I also made the interpreter reject keyword shadowing, which
broke some code on purpose and killed off a whole category of confusing bug. Two
weeks later I replaced Goldmark with my own Markdown renderer. Same API,
comparable speed, 1.5MB smaller. On a 10MB binary you notice 1.5MB.

**June: filling in the toolbox.** PostgreSQL joined MySQL under `sql()`. Image
processing became builtin. The linter shipped, and it checks Duso code inside
Markdown files too, so the docs get validated like source.

**July: the shape settles.** Regex became a real type, so strings stop quietly
compiling themselves into patterns behind your back. That broke code, and I did
it anyway, because the alternative was pattern injection in `find()`. The CLI
got rebuilt around subcommands. Data got encryption at rest. A WebSocket client
joined the server side. Discord and Slack modules landed. The interpreter got
2-4x faster at loops, calls, and allocations.

**August and September: out of single-node.** The unreleased 1.7 rebuilt the
write-ahead log into a proper op log, which finally made replication possible.
Leader and follower, streaming. Queue operations came out around 200x faster
and write about 2,000x less to disk. Process control grew up at the same time:
`exec()`, `shutdown()`, `schedule()`, and a `kill()` that actually kills.

## Key Metrics

*January 22 through September 5, 2026.*

| Thing | Number |
|---|---|
| Age | 7 months |
| Commits | 560 |
| Tagged builds | 426 |
| Formal releases | 10 |
| Issues closed | 130 |
| Issues open | 17 |
| Contributors | 1 |

## Built With Duso

There's no community yet. The Discord is quiet and every commit across all six
repos is mine. What I do have is a pile of things I built with it, because at
this point I don't reach for anything else.

**[arland.ai](https://arland.ai)** gives the runtime the hardest time. Arland is
an AI front-line sales rep, not a search box and not a support bot. It crawls a
customer's own pages, answers visitors out of them in whatever language they
wrote in, and mails the lead over instead of shoving a form in someone's face.
It drops onto someone else's site as one script tag. Behind that there's a
self-serve dashboard, billing, document ingestion, a public API, and the chat
itself, all Duso with HTMX on the front. No build step, no bundler, no
JavaScript toolchain.

It also uses most of the box. Six AI vendor modules sit behind the chat, OAuth
handles sign-in, Stripe handles payments, and the DigitalOcean module provisions
the servers customers get spun up on. Avatar uploads go through the built-in
image processing. WebSockets push live message-processing state to the browser.
Certbot renews TLS through the filtered `exec()`.

The data all lives in datastores, in memory for caching and encrypted
WAL-backed on disk for anything that has to survive a restart, split into one
datastore per customer bot. Tenant separation ends up being a naming decision
instead of an architecture. That's also why the 2.0 plan leaves the
`datastore()` API alone and only changes the mode: the seam that separates
tenants today is the one replication will run along later.

The traffic goes both ways. I shipped `exec()` on August 17 and TLS certificate
hot-reloading on the 18th, because something real needed certs to renew without
dropping connections. Most of 1.7 exists for reasons like that, which is the
closest thing this project has to a product manager.

**[shannan.dev](https://shannan.dev)** sells strategic market intelligence, the
kind of research that used to take months. Duso runs the AI agent harnesses
behind the reports and serves the site. They also sponsor Duso, so the tool
helps pay for itself.

**[duso.rocks](https://duso.rocks)** is the docs site, written in Duso,
generating the web manual out of the same docs embedded in the binary. It
documents itself with itself, which is either elegant or too clever, but it does
mean the published docs can't drift from the release.

**[balmer.dev](https://balmer.dev)** is my blog, built with Duso as a static or
dynamic site from the same source. It's also where the TODO app lives: full CRUD
with sessions and persistence in 164 lines of code and HTML, no dependencies,
compared against the same app in Node, Go, Python, Rails, PHP, and Rust.

Still cooking: a Discord bot server, and an iOS app that used to be called
Arland and is being remade as Arlee.

## Challenges

**It's just me.** 560 commits, one contributor, and a business now depending on
the thing. The bus factor is not great.

**I broke Homebrew and didn't notice for six weeks.** The formula worked fine
from February through June. Then my 1.6.21 update dropped the build suffix off
the download URLs, `v1.6.21` instead of `v1.6.21-512`, and every download
started 404ing. Two of the three `brew install` commands in my own docs pointed
at a tap that doesn't exist. So anyone who found Duso and ran the one-liner in
the README got an error for their trouble. Fixed in September. The flat download
numbers probably say more about that than about the language.

**The sandbox has holes.** Part of the pitch is running code an LLM wrote thirty
seconds ago with `-no-files` keeping it contained. Right now `require()` and
`include()` ignore that flag, and `fetch()` has no network restrictions at all.
1.8 is where I close that.

**Contrib kept falling behind.** Arland is where these modules actually get
used, so it's where they get fixed, and the fixes weren't making it home. The
Stripe module in Arland was 811 lines against 456 in contrib. A 705-line
DigitalOcean module wasn't in contrib at all. Both moved over this week. There
are still a few general-purpose helpers sitting in the app that belong in the
standard library.

**Replication is experimental.** It works end to end for what it covers, but I
shipped it unfinished on purpose and wrote a hardening plan against it.

**Rough edges.** `find()` and `replace()` are slow, especially with a predicate.
Two open panics. Windows install and the console spinner both need attention,
since I develop on a Mac and it shows.

## Looking Ahead

**1.8, make the sandbox real.** Close the `-no-files` holes, put file operations
behind `os.Root` so the kernel enforces the boundary, add a network allowlist
for `fetch()`, and set actual limits on processes, recursion, connections, and
datastore size. Fix the two panics. Get replication to stable.

**1.9, tools and speed.** Stepping and DAP support in the debugger so editors
can attach. Datastore indexes, since `select()` and `count()` scan everything
today. Sort out the `find()` and `replace()` slowness. A benchmark script that
spits out a report card on whatever machine you run it on.

**2.0, multi-node.** A distributed datastore with local, replicated, and remote
modes per namespace. The `datastore()` API stays put and the mode becomes a
deployment choice. 1.7's op log was the first move in that direction.

# Call for Contributors

If any of this sounds like something you want to exist, I really could use your
help. The issue tracker has a few things sized for a first contribution, and the
whole runtime is one Go binary you can build with `./build.sh`.

Start here:

- [duso.rocks](https://duso.rocks), the home page, with downloads and the full manual
- [README](/README.md), the quick tour
- [CONTRIBUTING.md](/CONTRIBUTING.md), how to send a patch
- [COMMUNITY.md](/COMMUNITY.md), how we treat each other
- [LICENSE](/LICENSE), Apache 2.0

The rest of the repos, if you want to poke at the editor tooling:

- [github.com/duso-org/duso](https://github.com/duso-org/duso), the runtime itself
- [github.com/duso-org/duso-vscode](https://github.com/duso-org/duso-vscode), VS Code extension
- [github.com/duso-org/duso-zed](https://github.com/duso-org/duso-zed), Zed extension
- [github.com/duso-org/duso-sublime](https://github.com/duso-org/duso-sublime), Sublime Text support
- [github.com/duso-org/duso-tree-sitter](https://github.com/duso-org/duso-tree-sitter), Tree-sitter grammar
- [github.com/duso-org/homebrew-duso](https://github.com/duso-org/homebrew-duso), the Homebrew formula
