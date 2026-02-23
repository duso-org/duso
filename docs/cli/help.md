# Usage

  duso [options]

## Options

### Learning Duso:
  `-read`              Browse files and docs (start here)
  `-docserver`         Start a local webserver with all docs

### Using Duso:
  `-init DIR`          Create a starter project in DIR
  `-doc TOPIC`         Display pretty docs for a module or builtin
  `-repl`              Start interactive REPL mode
  `-c `CODE``          Execute inline code
  `-debug`             Enable interactive debugger
  `-stdin-port PORT`   Use HTTP GET/POST in place of stdin/stdout

### Utility:
  `-install`           Install duso in your OS
  `-extract SRC DST`   Extract files from embedded filesystem to disk
  `-lsp`               Start an instance in LSP mode
  `-config OPTS`       Pass config as `key=value,key2=value` pairs
  `-lib-path PATH`     Pre-pend path to module search
  `-no-color`          Disable ANSI color output
  `-no-stdin`          Disable stdin (no waiting for input)
  `-help`              Show this help and exit
  `-version`           Show version and exit