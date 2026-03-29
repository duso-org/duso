# Usage

  `duso [options] script.du`   Run a script file (default if no command given)
  `duso [options] -repl`       Start interactive REPL mode
  `duso [options] -c 'CODE'`   Execute inline code

## Options

  `-config 'OPTS'`         Pass config as `key=value,key2=value` pairs
  `-debug`                 Enable interactive debugger
  `-lib-path PATH`         Pre-pend path to module search
  `-no-color`              Disable ANSI color output
  `-no-stdin`              Disable stdin (no waiting for input)
  `-no-files`              Disable local filesystem access
  `-stdin-port PORT`       Use HTTP GET/POST in place of stdin/stdout

# Utility Commands

  `duso -doc [TOPIC]`      Display pretty docs for a module or builtin
  `duso -docserver`        Start a local webserver with all docs
  `duso -read [PATH]`      Browse files and docs from embedded filesystem
  `duso -init DIR`         Create a starter project in DIR
  `duso -extract SRC DST`  Extract files from embedded filesystem to disk
  `duso -lsp`              Start LSP server on stdio
  `duso -lsp-tcp PORT`     Start LSP server on TCP port
  `duso -install`          Install duso binary to system PATH
  `duso -help`             Show this help and exit
  `duso -version`          Show version and exit
