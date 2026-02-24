# Documentation

## Getting Started

- [Learning Duso](docs/learning-duso.md)

## Guides

- [Files and Modules](docs/files-and-modules.md)
- [Debugging Scripts](docs/debugging-scripts.md)
- [Virtual Filesystem](docs/virtual-filesystem.md)
- [Distribution](docs/distribution.md)
- [Custom Distributions](docs/custom-distributions.md)
- [Internals](docs/internals.md)

## Libraries

### Standard Library (stdlib)

- [ansi](stdlib/ansi/ansi.md) - ANSI color and terminal styling utilities
- [markdown](stdlib/markdown/markdown.md) - Markdown parser with HTML and ANSI terminal output
- [docserver](stdlib/docserver/docserver.md) - Embedded documentation server with caching
- doccli - CLI documentation viewer

### Community Libraries (contrib)

- [claude](contrib/claude/claude.md) - Anthropic Claude API integration with multi-turn conversations and tools
- [couchdb](contrib/couchdb/couchdb.md) - CouchDB database client with CRUD operations and Mango queries
- [svgraph](contrib/svgraph/svgraph.md) - SVG chart and graph generation
- [zlm](contrib/zlm/zlm.md) - Zero Language Model for testing LLM-scale scenarios without burning tokens

## Language Keywords

- [if](docs/reference/if.md)
- [then](docs/reference/if.md)
- [else](docs/reference/if.md)
- [elseif](docs/reference/if.md)
- [end](docs/reference/if.md)
- [while](docs/reference/while.md)
- [do](docs/reference/for.md)
- [for](docs/reference/for.md)
- [in](docs/reference/for.md)
- [function](docs/reference/function.md)
- [return](docs/reference/return.md)
- [break](docs/reference/break.md)
- [continue](docs/reference/continue.md)
- [try](docs/reference/try.md)
- [catch](docs/reference/try.md)
- [and](docs/reference/if.md)
- [or](docs/reference/if.md)
- [not](docs/reference/if.md)
- [var](docs/reference/var.md)
- [raw](docs/reference/raw.md)
- [boolean](docs/reference/boolean.md)
- [nil](docs/reference/nil.md)

## Built-in Functions

### Strings

- [contains](docs/reference/contains.md)
- [find](docs/reference/find.md)
- [join](docs/reference/join.md)
- [len](docs/reference/len.md)
- [lower](docs/reference/lower.md)
- [repeat](docs/reference/string.md)
- [replace](docs/reference/replace.md)
- [split](docs/reference/split.md)
- [substr](docs/reference/substr.md)
- [template](docs/reference/template.md)
- [trim](docs/reference/trim.md)
- [upper](docs/reference/upper.md)

### Arrays & Objects

- [deep_copy](docs/reference/deep_copy.md)
- [filter](docs/reference/filter.md)
- [keys](docs/reference/keys.md)
- [map](docs/reference/map.md)
- [pop](docs/reference/pop.md)
- [push](docs/reference/push.md)
- [range](docs/reference/range.md)
- [reduce](docs/reference/reduce.md)
- [shift](docs/reference/shift.md)
- [sort](docs/reference/sort.md)
- [unshift](docs/reference/unshift.md)
- [values](docs/reference/values.md)

### Math

- [abs](docs/reference/abs.md)
- [ceil](docs/reference/ceil.md)
- [clamp](docs/reference/clamp.md)
- [floor](docs/reference/floor.md)
- [max](docs/reference/max.md)
- [min](docs/reference/min.md)
- [pow](docs/reference/pow.md)
- [random](docs/reference/number.md)
- [round](docs/reference/round.md)
- [sqrt](docs/reference/sqrt.md)
- [acos](docs/reference/number.md)
- [asin](docs/reference/number.md)
- [atan](docs/reference/number.md)
- [atan2](docs/reference/number.md)
- [sin](docs/reference/number.md)
- [cos](docs/reference/number.md)
- [tan](docs/reference/number.md)
- [exp](docs/reference/number.md)
- [log](docs/reference/number.md)
- [ln](docs/reference/number.md)
- [pi](docs/reference/number.md)

### File I/O

- [load](docs/reference/load.md)
- [save](docs/reference/save.md)
- [append_file](docs/reference/append_file.md)
- [copy_file](docs/reference/copy_file.md)
- [move_file](docs/reference/move_file.md)
- [rename_file](docs/reference/rename_file.md)
- [remove_file](docs/reference/remove_file.md)
- [list_dir](docs/reference/list_dir.md)
- [list_files](docs/reference/list_files.md)
- [make_dir](docs/reference/make_dir.md)
- [remove_dir](docs/reference/remove_dir.md)
- [file_exists](docs/reference/file_exists.md)
- [file_type](docs/reference/file_type.md)
- [current_dir](docs/reference/current_dir.md)

### Network & HTTP

- [fetch](docs/reference/fetch.md)
- [http_server](docs/reference/fetch.md)

### I/O

- [input](docs/reference/input.md)
- [print](docs/reference/print.md)
- [write](docs/reference/print.md)

### Date & Time

- [format_time](docs/reference/format_time.md)
- [now](docs/reference/now.md)
- [parse_time](docs/reference/parse_time.md)
- [sleep](docs/reference/sleep.md)

### JSON

- [format_json](docs/reference/format_json.md)
- [parse_json](docs/reference/parse_json.md)

### Modules

- [include](docs/reference/include.md)
- [require](docs/reference/require.md)

### Types

- [tobool](docs/reference/tobool.md)
- [tonumber](docs/reference/tonumber.md)
- [tostring](docs/reference/tostring.md)
- [type](docs/reference/type.md)

### Flow & Concurrency

- [context](docs/reference/context.md)
- [exit](docs/reference/exit.md)
- [parallel](docs/reference/parallel.md)
- [run](docs/reference/run.md)
- [spawn](docs/reference/spawn.md)

### Debugging

- [watch](docs/reference/watch.md)
- [throw](docs/reference/try.md)

### System

- [datastore](docs/reference/datastore.md)
- [doc](docs/reference/doc.md)
- [env](docs/reference/env.md)
- [uuid](docs/reference/uuid.md)

## Embedding

Learn about [Embedding Duso](docs/embedding/) in your Go applications.
