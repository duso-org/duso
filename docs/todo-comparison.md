# TODO App Comparison

The TODO MVP is a well-known way to show how a given language and framework handles a basic CRUD web app. We looked for TODO MVP examples that had a solid feature set to compare with Duso's example. We examine not just lines of code, but dependency exposure, memory footprint, operational burden, and learning curve.

## Duso Stands Out

- **Duso: 164 LOC, 0 dependencies, <1 minute setup, ~5MB runtime, zero admin**
- Other languages: 285–1,086 LOC, 1–300 dependencies, 12–25 minute setup, ~12–150MB runtime, ongoing admin

The difference isn't just code count. It's conceptual overhead, operational burden, and deployment footprint. Duso achieves parity with battle-tested frameworks while staying true to 90's-era simplicity: HTTP, templates, and a datastore. Modern runtime (goroutines, concurrent) underneath, but zero framework tax on top.

## Comparative Summary Table

| Metric | Duso | Node Express | Go Stdlib | PHP | Rust Rocket | Python Flask | Rails |
|--------|------|------|------|------|------|------|------|
| **LOC** | ==164== | 285 | 296 | 341 | 370 | 621 | 1,086 |
| **Dependencies** | ==0== | ~170 | ==1== | ==3== | ~300 | 12 | ~20 |
| **Persistence** | Built-in datastore | SQLite | MySQL | MySQL | SQLite | SQLite | PostgreSQL |
| **Setup Time** | ==<1 min== | 12 min | 15 min | 15 min (XAMPP, MySQL, schema init) | 20 min | 12 min | 25 min |
| **Binary/Deploy Size** | ==~12MB== | ~230MB | ~55MB | ~60MB | ==~8.5MB== | ~75MB | ~180MB |
| **Stack Runtime Memory** | ==~5MB== | ~100MB | ~105MB | ~130MB | ==~12MB== | ~40MB | ~150MB |
| **Framework Concepts** | ==2== | 6-8 | 6-8 | 4-5 | 8-10 | 8-10 | 12-15 |
| **CRUD** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Persistence** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Dynamic Interactions** | HTMX | None | None | jQuery AJAX | HTMX | None | None |
| **Tech Stack to Learn** | ==Duso, HTMX== | JavaScript, Node, Express, EJS, SQLite, SQL | Go, net/http, html/template, MySQL, SQL | PHP, MySQL, SQL, jQuery, Apache | Rust, Rocket, Handlebars, SQLx, SQLite, SQL, async/Tokio, HTMX | Python, Flask, Jinja2, SQLAlchemy, SQLite, SQL, HTMX | Ruby, Rails, ActiveRecord, ERB, PostgreSQL, SQL, migrations |
| **Admin Burden** | ==None== | Low | Medium (MySQL) | Medium-High (Apache, MySQL, jQuery CDN) | Low | Low | High (PostgreSQL) |

## Implementation Details by Language

### Duso (HTMX Example)

Built-in example [duso-org/duso/examples/htmx/todo](https://github.com/duso-org/duso/tree/main/examples/htmx/todo).

| Aspect | Details |
|--------|---------|
| **Language** | Duso (a less-quirky cousin of Lua) |
| **Framework** | Native `http_server()`, `datastore()` builtins |
| **Lines of Code** | **164 lines** across 9 modular `.du` files |
| **Persistence** | Built-in durable datastore backed by WAL |
| **Architecture** | Server-side template rendering + HTMX interactivity |
| **Session Model** | URL-based (`:session` path parameter) |

#### Dependencies
- **Runtime Framework:** None (duso builtins)
- **Database:** None (datastore primitive)
- **External Libraries:** None
- **Build Tools:** Run with `duso server.du`, or bundle app into its own binary

#### Operational Overhead
- **Setup Time:** <1 minute (download binary from duso.rocks, run)
- **Configuration:** Single `.du` file (19 lines routing)
- **Database Management:** None (datastore handles persistence + recovery)
- **Deployment:** Single 12MB self-contained binary (no runtime, no dependencies)

#### Learning Burden
- **Framework Concepts:** 2 (Duso language, HTMX for interactivity)
- **New Language:** Yes (Duso) — 1-2 hours for this example
- **Template System:** Native string interpolation
- **Persistence:** No SQL/schema knowledge needed

#### Admin Tasks
- None (no migrations, no schema versioning, no external service monitoring)

#### Features
- ✅ CRUD (Create, Read, Update, Delete)
- ✅ Toggle completion status
- ✅ Multi-session (per-URL isolation)
- ✅ Data persistence with recovery
- ✅ HTMX dynamic interactions
- ✅ Request isolation limits validation needs

### Node.js (Express + SQLite)

Simple Node.js implementation [jaredhanson/todos-express-sqlite](https://github.com/jaredhanson/todos-express-sqlite).

| Aspect | Details |
|--------|---------|
| **Language** | JavaScript (Node.js 12+) |
| **Framework** | Express.js 4.x |
| **Template Engine** | EJS (Embedded JavaScript) |
| **Lines of Code** | 285 lines |
| **Persistence** | SQLite (file-based, `db.sqlite3`) |
| **Architecture** | Server-side rendering with form-based interactions |
| **Session Model** | Request/form-based |

#### Dependencies
- **Runtime Framework:** Express.js, body-parser, method-override
- **Database:** SQLite (sqlite3 npm package)
- **Template Engine:** EJS
- **External Libraries:** 11 direct, ~170 total (npm dependency tree)
- **Build Tools:** npm, Node.js

#### Operational Overhead
- **Setup Time:** ~12 minutes (npm install, schema init)
- **Configuration:** Express middleware stack, SQLite path config
- **Database Management:** SQLite schema + migrations (via npm scripts)
- **Deployment:** Node.js binary (215MB) + node_modules (17MB) + app code (~230MB total)

#### Learning Burden
- **Framework Concepts:** 6-8 (Express routing, middleware, templates, request/response)
- **Dependencies:** 11 direct, ~170 transitive (npm dependency tree management)
- **New Language:** JavaScript (commonly known)
- **Template System:** EJS (JavaScript-based, familiar if you know JS)
- **Persistence:** SQLite, SQL basics, npm sqlite3 package

#### Admin Tasks
- Node.js version management (via nvm or similar)
- npm dependency updates (security patches)
- node_modules management (size, caching, lockfile)
- SQLite schema versioning

#### Features
- ✅ CRUD
- ✅ Form validation (HTML5 + basic server-side)
- ✅ Task persistence
- ✅ Vanilla CSS styling

### Go (Standard Library)

Minimal Go implementation [ichtrojan/go-todo](https://github.com/ichtrojan/go-todo).

| Aspect | Details |
|--------|---------|
| **Language** | Go 1.11+ |
| **Framework** | Go standard library (`net/http`, `html/template`) |
| **ORM/Query Layer** | MySQL driver (`database/sql`) |
| **Lines of Code** | 296 lines |
| **Persistence** | MySQL database |
| **Architecture** | MVC with standard library routing |
| **Session Model** | User-based (no explicit multi-session) |

#### Dependencies
- **Runtime Framework:** Go stdlib only
- **Database:** MySQL (external service)
- **Template Engine:** Go `html/template` (stdlib)
- **External Libraries:** `github.com/go-sql-driver/mysql`
- **Build Tools:** Go compiler, MySQL

#### Operational Overhead
- **Setup Time:** ~15 minutes (Go build, MySQL setup, schema init)
- **Configuration:** App config file, MySQL credentials, connection pooling
- **Database Management:** MySQL schema + migrations + backups + service monitoring
- **Deployment:** Single binary (~55MB total with MySQL)

#### Learning Burden
- **Framework Concepts:** 6-8 (routing, request/response, error handling)
- **New Language:** Go (if unfamiliar) — 4-6 hours for this example
- **Template System:** Go's html/template (stricter than Jinja2)
- **Persistence:** MySQL, SQL queries, database driver, connection pooling

#### Admin Tasks
- MySQL service management (running, monitoring, backups)
- Database schema versioning/migrations
- Connection pool tuning
- Replication (if scaling beyond single server)

#### Features
- ✅ CRUD
- ✅ Form validation (HTML5 + server-side)
- ✅ Task persistence
- ✅ Clean MVC structure

### PHP (Vanilla + MySQL + Apache)

Minimal PHP implementation [Smitha113/TODO-LIST](https://github.com/Smitha113/TODO-LIST).

| Aspect | Details |
|--------|---------|
| **Language** | PHP 7.0+ |
| **Framework** | None (vanilla PHP) |
| **ORM/Query Layer** | MySQLi (native PHP) |
| **Lines of Code** | 341 lines |
| **Persistence** | MySQL database |
| **Architecture** | Procedural, file-based routing, requires Apache web server |
| **Session Model** | jQuery AJAX, not HTMX |

#### Dependencies
- **Web Server:** Apache (required, can't run standalone)
- **Database:** MySQL (external service)
- **Runtime:** PHP (FastCGI via PHP-FPM)
- **Client-side:** jQuery (loaded from CDN)
- **Configuration:** Apache VirtualHost, PHP-FPM socket/FastCGI config
- **Build Tools:** XAMPP (Apache + MySQL bundle)

#### Operational Overhead
- **Setup Time:** ~15 minutes (install XAMPP, start services, run schema initialization via `localhost/index.php`, configure Apache document root)
- **Configuration:** Apache virtual host config, PHP-FPM pool config, MySQLi connection string, jQuery CDN inclusion
- **Database Management:** MySQL service management, schema creation, backups
- **Deployment:** Apache (~5MB) + PHP-FPM process (~25-40MB runtime) + MySQL (~50MB+) + application code

#### Learning Burden
- **Framework Concepts:** 4-5 (file-based routing, database connection, AJAX/jQuery, Apache configuration)
- **New Language:** No (PHP is ubiquitous)
- **Template System:** Inline PHP
- **Persistence:** MySQL + MySQLi driver + SQL
- **Operational:** Apache/PHP-FPM configuration, MySQL service management

#### Admin Tasks
- Apache web server management (startup, configuration, virtual hosts)
- PHP-FPM pool management (worker count, memory limits)
- MySQL service management (startup, backups, schema versioning)
- jQuery version management (if self-hosted)
- Web server monitoring and log management

#### Features
- ✅ Full CRUD (Create, Read, Update, Delete)
- ✅ Basic task management
- ✅ MySQL persistence
- ✅ Dynamic interactions (jQuery AJAX, no full page reloads)

#### Limitations
- ❌ Can't run standalone — requires Apache web server
- ❌ Requires MySQL service management (separate infrastructure)
- jQuery AJAX mechanism is heavier than HTMX
- More operational complexity: Apache + PHP-FPM + MySQL + jQuery versioning

### Rust (Rocket + HTMX + SQLite)

Architecturally closest to Duso [paultuckey/example-todo-app-rust-htmx](https://github.com/paultuckey/example-todo-app-rust-htmx).

| Aspect | Details |
|--------|---------|
| **Language** | Rust 1.56+ |
| **Framework** | Rocket web framework |
| **Query Layer** | SQLx (type-safe SQL, async) |
| **Template Engine** | Handlebars (server-side HTML) |
| **Lines of Code** | 370 lines |
| **Persistence** | SQLite (file-based) |
| **Architecture** | Server-side rendering + HTMX interactivity (identical to Duso) |
| **Session Model** | Request-based |

#### Dependencies
- **Runtime Framework:** Rocket, tokio (async runtime)
- **Database:** SQLite (sqlx async driver)
- **Template Engine:** Handlebars
- **External Libraries:** 4 direct, ~300 transitive (Cargo.lock: tokio, sqlx, handlebars, serde)
- **Build Tools:** Rust compiler, Cargo

#### Operational Overhead
- **Setup Time:** ~20 minutes (Rust setup, Cargo build, schema init)
- **Configuration:** Rocket.toml, SQLx database URL, async/await runtime config
- **Database Management:** SQLite schema + SQLx migrations (compile-time checked)
- **Deployment:** Single binary (~8.5MB release build) + .env config
- **Runtime Memory:** ~12MB (Rocket + SQLite embedded)

#### Learning Burden
- **Framework Concepts:** 8-10 (Rocket routing, async/await, futures, Handlebars, error handling)
- **New Language:** Rust (if unfamiliar) — 20-40 hours for this example (steep learning curve)
- **Template System:** Handlebars (similar to Jinja2 but stricter)
- **Persistence:** SQLite, SQL basics, type-safe SQL via SQLx (compile-time verification)
- **Async Concepts:** Futures, tokio runtime (additional mental model)

#### Admin Tasks
- Rust toolchain management (rustup)
- Dependency updates (Cargo.lock management, security audits)
- Compile-time checks (slower iteration during development)
- SQLite schema management

#### Features
- ✅ CRUD
- ✅ Dynamic HTMX interactions (same as Duso)
- ✅ Task persistence
- ✅ Type-safe SQL queries
- ✅ Async/concurrent request handling

### Python (Flask + HTMX + SQLAlchemy)

Feature-complete HTMX example [emarifer/flask-htmx-todolist](https://github.com/emarifer/flask-htmx-todolist).

| Aspect | Details |
|--------|---------|
| **Language** | Python 3.6+ |
| **Framework** | Flask 3.0+ |
| **ORM/Query Layer** | SQLAlchemy 2.0 + Flask-SQLAlchemy |
| **Lines of Code** | 621 lines (340 Python + 281 HTML) |
| **Persistence** | SQLite (file-based) |
| **Architecture** | Server-side rendering + HTMX interactivity |
| **Session Model** | User authentication with Flask-SQLAlchemy |

#### Dependencies
- **Runtime Framework:** Flask, Werkzeug
- **Database:** SQLite + Flask-SQLAlchemy + SQLAlchemy
- **Template Engine:** Jinja2
- **External Libraries:** 12 direct, minimal transitive
- **Build Tools:** pip, virtualenv

#### Operational Overhead
- **Setup Time:** ~12 minutes (venv, pip install, DB init)
- **Configuration:** Flask app factory, SQLAlchemy config, environment setup
- **Database Management:** SQLAlchemy ORM handles schema, manual migrations possible
- **Deployment:** Python runtime (~30MB) + packages (~12MB) + app (~1MB) = ~75MB

#### Learning Burden
- **Framework Concepts:** 8-10 (Flask routing, SQLAlchemy ORM, authentication, HTMX, Jinja2 templates)
- **New Language:** No (Python commonly known)
- **Template System:** Jinja2 (with HTMX attributes)
- **Persistence:** SQLAlchemy ORM, SQL basics, SQLite

#### Admin Tasks
- Database schema management (SQLAlchemy handles much of this)
- User/session management
- Python version management
- Virtual environment maintenance
- Dependency updates

#### Features
- ✅ Full CRUD (Create, Read, Update, Delete)
- ✅ User authentication
- ✅ Multi-user support (per-user task isolation)
- ✅ Task editing and updates
- ✅ HTMX dynamic interactions (no page reloads)
- ✅ Form validation

### Ruby on Rails (With Authentication)

Enhanced version with user auth [drmikeh/rails_todo_app](https://github.com/drmikeh/rails_todo_app).

| Aspect | Details |
|--------|---------|
| **Language** | Ruby 2.2+ |
| **Framework** | Rails 4.2+ |
| **ORM** | ActiveRecord |
| **Lines of Code** | 1,086 lines (excluding migrations) |
| **Persistence** | PostgreSQL 9.4+ (configurable to SQLite) |
| **Architecture** | Classical MVC with convention-over-configuration |
| **Session Model** | Built-in session management |

#### Dependencies
- **Runtime Framework:** Rails (ActionPack, ActiveRecord, ActionView, etc.)
- **Database:** PostgreSQL (or SQLite, MySQL)
- **ORM:** ActiveRecord
- **External Libraries:** ~20 gems (Gemfile.lock)
- **Build Tools:** Ruby, Bundler, gem package manager

#### Operational Overhead
- **Setup Time:** ~25 minutes (Ruby/Rails setup, bundle install, DB init, schema migration)
- **Configuration:** database.yml, secrets.yml, multiple environment configs
- **Database Management:** PostgreSQL/MySQL schema + ActiveRecord migrations + rake tasks
- **Deployment:** Ruby runtime (~50MB) + gems (~100MB) + database (~30MB) = ~180MB

#### Learning Burden
- **Framework Concepts:** 12-15 (MVC, routing, models, views, controllers, migrations, associations, validations)
- **New Language:** Ruby (if unfamiliar) — 10-15 hours for basics, 40+ hours for Rails idioms
- **Template System:** ERB (Embedded Ruby)
- **Persistence:** PostgreSQL, SQL basics, ActiveRecord ORM, migrations
- **Conventions:** Rails has many implicit behaviors (schema naming, foreign keys, etc.)

#### Admin Tasks
- Database migrations (version control, backwards compatibility)
- Gem dependency management (security patches, version conflicts)
- Ruby version management (rbenv, rvm)
- Database backups/restore (PostgreSQL complexity)

#### Features
- ✅ CRUD
- ✅ Full validation layer (ActiveRecord validations)
- ✅ Task persistence
- ✅ Session management (built-in)

**Note:** This example includes authentication and per-user task lists by default, making it more feature-complete than typical TODO MVPs. Simpler Rails examples (e.g., `iridakos/todo`) skip authentication entirely.

## Key Insights

### Dependency "Haze"
- **Duso:** 0 external dependencies. Builtins provide everything.
- **Go stdlib:** 1 database driver, minimal ecosystem.
- **PHP:** 3 (MySQL, Apache, jQuery).
- **Python Flask:** 12 direct, minimal transitive.
- **Rails:** ~20 gems, modest transitive dependencies.
- **Node Express:** 11 direct, ~170 transitive (npm dependency tree).
- **Rust Rocket:** 4 direct, ~300 transitive (Cargo.lock).

### Operational Complexity Curve
1. **Duso & Rust:** Flat. Single binary, embedded database, no service monitoring.
2. **Flask:** Slight slope. Local venv, minimal dependencies, schema management.
3. **PHP & Go:** Medium. External MySQL service management required, schema versioning, configuration.
4. **Node & Rails:** Steep. Massive dependency trees (170+ for Node, 20+ for Rails), service management, migrations, version conflicts.

### Learning Burden Beyond LOC
- **Duso:** LOC matches learning (short language intro, then native APIs)
- **Python Flask:** LOC > learning (ORM, SQLAlchemy patterns beyond the code)
- **Go:** LOC > learning (SQL + driver + MySQL service concepts add cognitive load)
- **Node:** LOC << learning (170 transitive deps, npm ecosystem, async patterns)
- **Rust:** LOC < learning (type system, async, lifetimes, tokio runtime)
- **Rails:** LOC << learning (conventions, magic, 20+ gems, Active Record patterns)

## Duso is Built for Simplicity

The costs of other choices aren't "wrong". They all come down to tradeoffs. Flask adds an ORM option. Rails adds conventions and migrations. Node brings an ecosystem. Rust enforces type safety. Each buys you something and each costs you something. Duso trades a lot of niche features and complexities for its simplicity, but doesn't sacrifice power (thanks to Go). 

Duso is built for the majority of servers that run great on a single node, while providing a smooth and approachable developer experience.
