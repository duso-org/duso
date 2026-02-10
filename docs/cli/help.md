# Usage

  duso [options] <script.du>

## Options

  `-init DIR`          Create a starter project in DIR
  `-repl`              Start interactive REPL mode
  `-c CODE`            Execute inline code

  `-doc TOPIC`         Display docs for a module or builtin
  `-docserver`         Start a local webserver with all docs
  `-lsp`               Start an instance in LSP mode

  `-debug`             Enable interactive debugger with breakpoint()
  `-stdin-port PORT`   HTTP transport for stdin/stdout via HTTP API
                       (enables LLM and remote interaction)

  `-config OPTS`       Pass configuration as `key=value` pairs
  `-lib-path PATH`     Pre-pend path to module search
  `-no-color`          Disable ANSI color output
  `-no-stdin`          Disable stdin (no waiting for input)
  `-help`              Show this help and exit
  `-version`           Show version and exit
