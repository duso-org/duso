`duso script.du [OPT]`        Run a script file
`duso debug script.du [OPT]`  Run script with debugger
`duso lint a.du b.md [OPT]`   Validate code in script or markdown files
`duso eval 'CODE' [OPT]`      Execute inline code
`duso repl [OPT]`             Start interactive REPL

`duso primer`                 Display plain text primer for coding agents
`duso init DIR`               Create starter project in DIR
`duso doc [TOPIC]`            Display ref doc for keyword, built-in, or type
`duso read [PATH]`            Browse embedded docs, examples, and other files
`duso webdoc [OPT]`           Start local web server for docs
`duso version`                Show version

`duso lsp [-port PORT]`       Start LSP server w stdin/out or tcp/ip on PORT
`duso extract SRC DST`        Extract embedded files to disk
`duso syntax`                 Generate syntax config for editor plugins
`duso install`                Install duso binary to system PATH

## options

`-config 'k=num,k="str"'`     Pass config to script
`-lib-path PATH`              Prepend PATH to module search
`-no-color`                   Disable ANSI colors
`-no-stdin`                   Disable stdin
`-no-files`                   Disable filesystem access
`-stdin-port PORT`            HTTP transport for stdin/stdout
`-ignore-warnings`            Suppress non-error diagnostics (lint)
