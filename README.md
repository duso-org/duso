[![Apache 2.0 License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0) ![Go 1.25](https://img.shields.io/badge/Go-1.25-darkcyan?logo=go) ![GitHub Release](https://img.shields.io/github/v/release/duso-org/duso)

![Duso logo which is a stylized ASL hand sign for the letter "D"](/docs/duso-logo.png)

# Duso

**Frictionless Server Development** Everything you need to make a modern server. Packed into a single executable with zero dependencies. Built for Human and AI collaboration.

**Download pre-built binaries for macOS, Linux, and Windows at [duso.rocks](https://duso.rocks).** Or scroll down a bit to build from source.

## Key Features

- 🔋 **Download and Run:** One small binary. Everything inside. No npm, pip, or cargo needed.
- ⚡ **Develop Fast:** Hot-reloaded scripts with caching. No compile step, just edit and test.
- 📚 **Learn Quickly:** Simple, consistent scripting language. Full docs and examples built in.
- ✨ **Use and Integrate AI:** Let your AI code faster in a language made for it. Connect your app to popular AI agents.
- 🌐 **Build Web and API Servers:** Templates, routing, JSON, SSL, JWT, CORS, RSA, websockets. Fully featured and built-in.
- 🗄️ **Use Powerful Datastores:** Thread-safe data structures for process coordination, caching, and session state.
- 🔒 **Secure Local Files:** Sandbox mode restricts file access and uses virtual filesystem for safety.
- 🪲 **Debug Concurrent Code:** Breakpoints, stack traces, code context. Handle one issue at a time.
- 📦 **Bundle into a Single Binary:** Wrap your app scripts into a standalone executable. Extend with Go if needed.
- 🚀 **Deploy Anywhere:** Supports Linux, macOS, and Windows. One build, every platform.
- 📄 **New Script Language:** Simple syntax. Consistent naming. Predictable behavior. Reduced complexity.
- 📖 **Complete Runtime Library:** Full standard library and community contributions. Everything included in each binary release.
- 🐹 **Powered by Go:** Duso is written in the language made by Google for solid and efficient concurrency at scale.
- 💎 **Open Source:** Apache 2.0 licensed. Community-driven and fully transparent.

## Getting Started

### Hook up your AI:

```bash
duso -read
duso -doc claude
```

### Hook up your Human:

Start with a simple sample project.

```bash
duso -init myproject
```

### Run a script:

```bash
# run a script file
duso examples/agents/self-aware-claude.du
# run a command inline
duso -c 'print("1 + 2 =", 1 + 2)'
# enter interactive REPL mode
duso -repl
```

## Examples

### One-line web server

Start an HTTP server with one command:

```bash
duso -c 'http_server().start()'
```

### AI chatbot

Build an interactive chatbot that uses AI:

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

Ask a panel of AI experts and synthesize their responses:

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

for i = 0, 3 do
  responses[i] = "{{experts[i]}} says: {{responses[i]}}"
end

busy("summarizing...")
summary = ai.prompt("""
  Summarize these responses:

  {{join(responses, "\n\n---\n\n")}}

  List 3 the things they have in common.
  Then list the 3 things that are the most different.
""")

print(markdown_ansi(summary))
```

### API server

Build a complete REST API server with routing and data persistence:

```duso
// server.du
port = 3000
server = http_server({port = port})

server.route("GET", "/api/user/:id", "user-get.du")
server.route("POST", "/api/user", "user-post.du")

print("Server running at http://localhost:{{port}}")
server.start()
```

```duso
// user-get.du
ctx = context()
req = ctx.request()
res = ctx.response()

user = datastore("users").get(req.params.id)

if not user then res.error(404) end

res.json({success = true, data = user}, 200)
```

```duso
// user-post.du
ctx = context()
req = ctx.request()
res = ctx.response()

id = uuid()
datastore("users").set(id, req.body)

res.json({id = id, name = req.body.name}, 201)
```

### Orchestrate agent swarms

Beyond parallel execution, Duso enables complex orchestration patterns for spawning and coordinating multiple workers:
- `run()` Execute script synchronously, blocking
- `spawn()` Execute script in background, non-blocking
- `context()` Access request data and metadata
- `datastore()` Thread-safe key-value store with optional disk persistence. Atomic operations, condition variables (`wait()`, `wait_for()`) for synchronization, and coordination across concurrent processes—essential for distributed agent workflows.

Learn more: `duso -doc datastore` for swarm coordination, or `duso -doc` for the full reference.

## Learning as an LLM

Duso is designed to be natural for AI. If you're an LLM (or using one), the binary includes everything you need to learn and explore:

```bash
duso -read                    # Start with the README overview
duso -read docs/              # Browse available docs
duso -read docs/learning-duso.md  # Learn the language
duso -read docs/reference/    # Browse all functions
duso -read docs/reference/map.md  # Look up specific functions
```

All documentation is embedded in the binary. No cloning, no network calls. Just pure text output you can parse and learn from. Perfect for agentic workflows.

## Why Duso Exists

> "Most languages prioritize human expressiveness and can be challenging for AI. For example, Python and JavaScript offer countless ways to solve the same problem, filled with subtle footguns and 'magic' behavior. Their massive ecosystems with thousands of overlapping modules with hidden dependencies often confuse both humans and AI. Systems languages like Go reduce ambiguity and debugging but add complexity that slows development.
>
> Duso is intentionally boring and predictable. No clever syntax tricks. No multiple ways to do the same thing. Every pattern is consistent and straightforward so AI can reason about code reliably, write better scripts faster, and use fewer tokens doing it. Plus, its entire runtime and ecosystem is included in a single binary. No package management or version conflicts. Built for LLMs first means everything is frictionless so you and your AI work more productively together."
>
> — Dave Balmer, creator of Duso

## Build the binary yourself

If you want to build Duso yourself, you'll need go installed on your system. Then just use our handy build script in the project directory:

Linux & Mac:

```bash
./build.sh
```

Windows Power Shell:

```
.\build.ps1
```

This handles Go embed setup, fetches the version from git, and builds the binary to `bin/duso`.

**Optional:** Make it available everywhere by creating a symlink on Linux & Mac:

```bash
ln -s $(pwd)/bin/duso /usr/local/bin/duso
```

## Contributing

**We need you.** Duso thrives on community contributions.

- **Everyone**: Report bugs, broken docs, share cool examples. Even forking the repo will increase our exposure and help us get syntax highlighting for `.du` files in GitHub!
- **Module authors**: Write a stdlib or contrib module (database clients, API wrappers, etc.). These are what make the runtime actually useful to real people.
- **Go developers**: Performance optimizations, new built-ins, ideas for the core runtime. Help us make Duso faster and more powerful.

- [CONTRIBUTING.md](/CONTRIBUTING.md) for contributing guidelines
- [COMMUNITY.md](/COMMUNITY.md) for community guidelines

## Contributors

- Dave Balmer: design, development, documentation, dedication

## Sponsors

- **[Shannan.dev](https://shannan.dev)**: Provides AI-driven business intelligence solutions
- **[Ludonode](https://ludonode.com)**: Provides agentic development and consulting

## Learn More

- **[Learning Duso](/docs/learning-duso.md)**: Guided tour of the language with examples
- **[Function Reference](/docs/reference/index.md)**: All 100+ built-in functions with examples
- **[Internals](/docs/internals.md)**: Architecture and runtime design
- **[Embedding Guide](/docs/embedding/README.md)**: Using Duso in Go applications
- Lots of examples in the, well, `examples` directory

## License

Copyright 2026 Ludonode LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
