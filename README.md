[![Apache 2.0 License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) ![Go 1.25](https://img.shields.io/badge/Go-1.25-darkcyan?logo=go) ![GitHub Release](https://img.shields.io/github/v/release/duso-org/duso)

![Duso the cat mascot](/docs/duso-cat-github.png)

# Duso

**One 10MB binary. A whole server stack.**

Duso is a scripting language and server runtime in a single Go binary. HTTP server, ACID datastore, WebSockets, AI clients, image processing, a debugger, a linter, even the documentation — all built in. No npm. No virtualenv. No stack to assemble before you can write line one. And it idles at about 5MB of RAM, so it runs happily on the cheapest VPS you can rent.

```bash
duso -c 'http_server().start()'
```

That's a web server. It's already running on `http://localhost:8080`.

## Download & Install

**Pre-built binaries:** [duso.rocks/download](https://duso.rocks/download)

Or install with Homebrew:

```bash
brew install duso-org/tap/duso
```

Or build from source:

**Linux & macOS:**
```bash
git clone https://github.com/duso-org/duso.git
cd duso
./build.sh
```

**Windows PowerShell:**
```powershell
git clone https://github.com/duso-org/duso.git
cd duso
.\build.ps1
```

Then optionally symlink it (Linux & macOS):
```bash
ln -s $(pwd)/bin/duso /usr/local/bin/duso
```

## What Makes Duso Different

**There's no async/await.** Duso doesn't have colored functions. `fetch()` just returns the response. When you actually want concurrency, `spawn()` a script or run functions in `parallel()` — real goroutines underneath, with closure safety handled for you. You get Go's concurrency without writing Go, and without the `await` confetti.

**Hot reload isn't a build tool.** Route handlers are plain scripts, checked for changes on every request. Save the file, hit refresh. There is no watcher to configure, no bundler, no restart. Dev and prod run the same binary the same way.

**The datastore replaces Redis for most apps.** An ACID key-value store lives in the runtime — in-memory or WAL-persisted to disk, with atomic increments, blocking waits, and real queues (`push()`, `pop_wait()`). Spawned workers coordinate through it without you setting up a single external service.

**The docs are in the binary.** `duso doc <anything>` works on a plane. `duso read` gives you a guided tour, `duso webdoc` serves the whole manual locally, and `duso extract examples ./examples` hands you runnable code.

**Untrusted code can be sandboxed.** Run a script with `-no-files` and it's confined to virtual filesystems (`/EMBED/` read-only, `/STORE/` backed by the datastore) with no real filesystem or environment access. Handy when the script was written thirty seconds ago by an LLM.

**Your app can become its own binary.** `bundle-duso` embeds your scripts, configs, and static files into a standalone executable — cross-compiled for Linux, macOS, or Windows. Deployment is `scp` and run.

## Quick Start

### Run a script

```bash
duso script.du
```

### Run inline code

```bash
duso -c 'print("Hello, World!")'
```

### Interactive REPL mode

```bash
duso repl
```

### Basic AI prompt

The AI clients are built in — no SDK to install, no wrapper library to pick:

```duso
ai = require("claude")
print(ai.prompt("What is 2+2?"))
```

### Interactive chatbot

```duso
ai = require("openai")
chat = ai.session()

while true do
  prompt = input("\n\nYou: ")
  if lower(prompt) == "exit" then break end

  write("\n\nOpenAI: ")
  busy("thinking...")
  write(chat.prompt(prompt))
end
```

### AI workflow with parallel experts

Four experts answer at once — this is `parallel()` doing real concurrent work, and those `{{expr}}` string templates are native syntax:

```duso
ai = require("claude")

prompt = input("Ask the panel: ")
busy("asking...")

experts = ["Astronomer", "Astrologer", "Biologist", "Accountant"]
responses = parallel(map(experts, function(expert)
  return function()
    return ai.prompt(prompt, {
      system = """
        You are an expert {{expert}}. Always reason and
        interact from this mindset. Limit your field of
        knowledge to this expertise.
      """,
      max_tokens = 500
    })
  end
end))

busy("summarizing...")
summary = ai.prompt("""
  Summarize these responses:
  {{join(responses, "\n\n---\n\n")}}

  List 3 things they have in common.
  Then list the 3 things that are the most different.
""")

print(markdown_ansi(summary))
```

### REST API with database

Each handler is its own script and runs in its own goroutine. Edit one while the server is running and the next request picks it up:

```duso
// server.du
server = http_server({port = 3000})

server.route("GET", "/users/:id", "get-user.du")
server.route("POST", "/users", "create-user.du")

print("Running on :3000")
server.start()
```

```duso
// get-user.du
ctx = context()
user = datastore("users").get(ctx.request().params.id)
ctx.response().json(user)
```

```duso
// create-user.du
ctx = context()
user = ctx.request().json()
datastore("users").set(user.id, user)
ctx.response().json(user)
```

### Concurrent workers with shared state

Ten workers, shared counters, no mutexes and no message broker:

```duso
// bees.du - Spawn workers, wait for all to finish
bees = 10
swarm = datastore("swarm")
swarm.set("done", 0)
swarm.set("buzzes", 0)

for i = 1, bees do
  spawn("worker.du", {bee_id = i})
end

swarm.wait("done", bees)
print("All done! Total buzzes: " + swarm.get("buzzes"))
```

```duso
// worker.du - Spawned worker increments shared counters
ctx = context()
swarm = datastore("swarm")

buzzes = ceil(random() * 10)
for i = 1, buzzes do
  sleep(random() * 0.5)
  swarm.increment("buzzes")
end

swarm.increment("done")
```

## The Language

The whole design is deliberately boring in the best way: one obvious way to do each thing, no metaprogramming rabbit holes, no operator overloading surprises. That predictability is also why AI assistants write good Duso on the first try — there's no trivia to trip over.

## Also in the Box

- **Full web stack**: routing, templates, static files, WebSockets, SSL, CORS, JWT
- **SQL databases**: PostgreSQL, MySQL, MariaDB, TiDB, CockroachDB — one `sql()` builtin
- **AI integrations**: Claude, OpenAI, Gemini, Groq, Ollama, Azure AI, DeepSeek
- **Image processing**: load, crop, scale, rotate, composite; PNG, JPEG, GIF
- **Everyday tools**: HTTP client, crypto, base64, UUID, date/time, Markdown rendering
- **Integrated debugger**: breakpoints, stack traces, concurrent-aware
- **Editor support**: built-in LSP server plus extensions for VS Code, JetBrains, Vim
- **Starter kit**: `duso init myproject` scaffolds a working project

## Why I Made Duso

**Duso is intentionally simple and predictable.** No magic. No multiple ways to do the same thing. Every pattern is consistent so AI can reason about code reliably and write better scripts faster.

**Duso is a joy to use.** Everything including the runtime, libs, and docs is bundled in a single 10MB binary. No package management. No version conflicts. No stack building. Duso makes coding fun again.

[Dave Balmer](https://balmer.dev), creator of Duso

## Full Documentation

- **Website:** [duso.rocks](https://duso.rocks)
- **Learning Guide:** [docs/learning-duso.md](/docs/learning-duso.md)
- **Built-in:** `duso read` or `duso doc <topic>`
- **Discord:** [Join the community](https://discord.gg/aecPVqmsW7)

## Community Libraries

Built-in integrations for:

- **AI:** Claude, OpenAI, Gemini, Grok, Groq, Ollama, Azure AI, DeepSeek
- **Messaging:** Discord, Slack
- **Databases:** CouchDB (SQL databases are handled by the built-in `sql()` — no module needed)
- **Payments:** Stripe
- **Testing & Utils:** Icons (Phospher), SVG graphs (svgraph), Zero Language Model (zlm)

See [contrib/](/contrib/) for full list and docs.

## Contributing

Duso is open source under Apache 2.0. Contributions welcome:

1. Fork [github.com/duso-org/duso](https://github.com/duso-org/duso)
2. Create a branch (`git checkout -b feature/thing`)
3. Commit changes (`git commit -am 'add thing'`)
4. Push to branch (`git push origin feature/thing`)
5. Open a Pull Request

See [CONTRIBUTING.md](/CONTRIBUTING.md) and [COMMUNITY.md](/COMMUNITY.md) for guidelines.

## Contributors

- [Dave Balmer](https://balmer.dev): design, development, documentation, dedication

## Sponsors

- **[Shannan.dev](https://shannan.dev)**: Provides AI-driven business intelligence solutions
- **[Ludonode](https://ludonode.com)**: Provides agentic development and consulting

## FAQ

**Q: Is Duso production-ready?**
A: Yes.

**Q: How do I deploy Duso?**
A: It's one binary. `scp` it to a server and run it. Alternatively, containerize it in Docker or deploy to Fly.io, Railway, or Heroku.

**Q: Can I bundle scripts into the binary?**
A: Yes. Use `bundle-duso` to build a standalone executable with your scripts, configs, and static files embedded. See [docs/bundling-applications.md](/docs/bundling-applications.md).

**Q: Can I extend Duso with Go?**
A: Yes. The language is designed to be extended. You can write custom builtins in Go.

## License

Apache 2.0. See [LICENSE](/LICENSE).
