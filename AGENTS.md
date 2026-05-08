# PROJECT KNOWLEDGE BASE

**Generated:** 2026-03-23
**Commit:** 82cd224
**Branch:** master

## OVERVIEW

Go configuration library (`github.com/beatlabs/harvester`) for dynamic config seeding and monitoring via struct tags. Sources: seed values, env vars, CLI flags, files, Consul, Redis. Uses concurrent-safe `sync` types and a builder pattern with functional options.

## STRUCTURE

```
harvester/
├── harvester.go       # Public API: New(), Harvester interface, Harvest flow
├── options.go         # Functional options: WithConsul*, WithRedis*
├── doc.go             # Package doc
├── config/            # Struct tag parsing, Field/Config types, CfgType interface
│   ├── config.go      # Field, Config, ChangeNotification, Source constants
│   └── parser.go      # Reflection-based struct parser, duplicate detection
├── seed/              # Seeding phase: applies values from all sources
│   ├── seed.go        # Seeder, Getter interface, source processing chain
│   ├── consul/        # Consul getter (KV API)
│   └── redis/         # Redis getter
├── monitor/           # Monitoring phase: watches for runtime changes
│   ├── monitor.go     # Monitor, Watcher interface, change application
│   ├── consul/        # Consul watcher (watch.Plan for keys/prefixes)
│   └── redis/         # Redis watcher (polling with hash-based change detection)
├── sync/              # Concurrent-safe config value types
│   ├── generic.go     # Value[T] — generic RWMutex-protected container
│   └── sync.go        # Bool, Int64, Float64, String, Secret, TimeDuration, Regexp, StringMap, StringSlice
├── change/            # Change DTO (source, key, value, version)
├── examples/          # Runnable example (main.go)
└── scripts/           # gofmtcheck.sh
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add config source | `config/config.go` (Source const), `seed/seed.go` (process* func), `options.go` (WithX option) | Follow existing seed→env→file→consul→redis→flag chain |
| Add sync type | `sync/sync.go` | Must implement `config.CfgType` (String() + SetString()) |
| Add monitor backend | `monitor/` subpackage | Implement `monitor.Watcher` interface |
| Add seed backend | `seed/` subpackage | Implement `seed.Getter` interface |
| Change notification | `config/config.go` `Field.Set()` → `sendNotification()` | Version-gated: older versions silently ignored |
| Nested config structs | `config/parser.go` `getFields()` | Recursion with prefix concatenation |
| Integration tests | `*_integration_test.go` | Build tag `//go:build integration`, need Consul+Redis running |
| Unit tests | `*_test.go` (co-located) | Table-driven with `testify/assert` + `testify/require` |

## ARCHITECTURE

```
User struct (tags) → config.New() → parser → []*Field
                                                 ↓
harvester.New(cfg, ch, opts...) → Seeder.Seed() → seed→env→file→consul→redis→flags
                                       ↓
                                  Monitor.Monitor(ctx) → Watcher.Watch() → chan []*Change → Field.Set()
```

- **Seeding order**: seed tag → env → file → consul → redis → CLI flags (last wins per source)
- **Versioning**: Consul uses `ModifyIndex`, Redis uses hash-based change detection with synthetic versions
- **Notifications**: Optional `chan<- config.ChangeNotification` passed to `New()`

## KEY INTERFACES

| Interface | Package | Method | Implementors |
|-----------|---------|--------|-------------|
| `Harvester` | root | `Harvest(ctx) error` | `harvester` struct |
| `Seeder` | root | `Seed(*config.Config) error` | `seed.Seeder` |
| `Monitor` | root | `Monitor(ctx) error` | `monitor.Monitor` |
| `Getter` | seed | `Get(key) (*string, uint64, error)` | `seed/consul.Getter`, `seed/redis.Getter` |
| `Watcher` | monitor | `Watch(ctx, chan<- []*change.Change) error` | `monitor/consul.Watcher`, `monitor/redis.Watcher` |
| `CfgType` | config | `String() string`, `SetString(string) error` | All `sync.*` types |

## CONVENTIONS

- **Vendor mode**: dependencies vendored, `go mod vendor` used. golangci-lint runs with `-mod=vendor`
- **Logging**: `log/slog` throughout (no third-party logger)
- **Error handling**: Return errors up, log non-fatal issues (missing consul key, missing env var) and continue
- **Struct tags**: `seed`, `env`, `flag`, `file`, `consul`, `redis` — parsed in `config/config.go` `sourceTags` array
- **Duplicate detection**: Consul and Redis keys checked for duplicates in parser
- **Concurrency**: All config field reads/writes via `sync.Value[T]` (RWMutex). `Field.Set()` has its own mutex for version gating
- **Formatters**: gofmt, gofumpt, goimports enforced via golangci-lint

## ANTI-PATTERNS (THIS PROJECT)

- **Never use non-sync types** in config structs — must implement `CfgType` interface
- **Never skip version check** — `Field.Set()` silently rejects older versions (by design)
- **No binary files** with `file` source tag — text only
- **No duplicate Consul/Redis keys** across fields — parser rejects them
- **Secret fields** always display as `***` in logs/String() — do not circumvent

## COMMANDS

```bash
# Dev commands (requires: go-task v3, Docker)
task test              # Unit tests: go test ./... -cover -race
task testint           # Integration tests (needs deps): -tags=integration
task ci                # CI suite: integration + coverage profile
task lint              # Lint via Docker: golangci-lint (vendor mode)
task fmt               # Format: go fmt ./...
task fmtcheck          # Check formatting (CI gate)
task deps-start        # Start Consul + Redis: docker compose up -d
task deps-stop         # Stop deps: docker compose down

# Direct Go commands
go test ./... -cover -race                    # Unit tests only
go test ./... -tags=integration -cover -race  # All tests (deps must be running)
```

## NOTES

- **Go version**: 1.26.1 (from go.mod)
- **CI**: GitHub Actions on push to master + PRs. Two jobs: lint + build (integration tests with docker-compose deps)
- **Coverage**: Uploaded to Codecov via `coverage.txt`
- **PRs**: Require 2 curator approvals, squash-merged. Follow CONTRIBUTE.md
- **DCO**: Commits must be signed off (`Signed-off-by:`)
- **Code owners**: @mantzas @pkritiotis
- **Consul watcher**: Uses `hashicorp/consul/api/watch.Plan` with `RunWithClientAndHclog` — custom hclog→slog adapter in `monitor/consul/log.go`
- **Redis watcher**: Polling-based with `MGET` + SHA256 hash comparison for change detection
- **Flag parsing**: Filters `os.Args` to only harvester-defined flags, discards unknown flags silently

---

# context-mode — MANDATORY routing rules

You have context-mode MCP tools available. These rules are NOT optional — they protect your context window from flooding. A single unrouted command can dump 56 KB into context and waste the entire session.

## BLOCKED commands — do NOT attempt these

### curl / wget — BLOCKED
Any shell command containing `curl` or `wget` will be intercepted and blocked by the context-mode plugin. Do NOT retry.
Instead use:
- `context-mode_ctx_fetch_and_index(url, source)` to fetch and index web pages
- `context-mode_ctx_execute(language: "javascript", code: "const r = await fetch(...)")` to run HTTP calls in sandbox

### Inline HTTP — BLOCKED
Any shell command containing `fetch('http`, `requests.get(`, `requests.post(`, `http.get(`, or `http.request(` will be intercepted and blocked. Do NOT retry with shell.
Instead use:
- `context-mode_ctx_execute(language, code)` to run HTTP calls in sandbox — only stdout enters context

### Direct web fetching — BLOCKED
Do NOT use any direct URL fetching tool. Use the sandbox equivalent.
Instead use:
- `context-mode_ctx_fetch_and_index(url, source)` then `context-mode_ctx_search(queries)` to query the indexed content

## REDIRECTED tools — use sandbox equivalents

### Shell (>20 lines output)
Shell is ONLY for: `git`, `mkdir`, `rm`, `mv`, `cd`, `ls`, `npm install`, `pip install`, and other short-output commands.
For everything else, use:
- `context-mode_ctx_batch_execute(commands, queries)` — run multiple commands + search in ONE call
- `context-mode_ctx_execute(language: "shell", code: "...")` — run in sandbox, only stdout enters context

### File reading (for analysis)
If you are reading a file to **edit** it → reading is correct (edit needs content in context).
If you are reading to **analyze, explore, or summarize** → use `context-mode_ctx_execute_file(path, language, code)` instead. Only your printed summary enters context.

### grep / search (large results)
Search results can flood context. Use `context-mode_ctx_execute(language: "shell", code: "grep ...")` to run searches in sandbox. Only your printed summary enters context.

## Tool selection hierarchy

1. **GATHER**: `context-mode_ctx_batch_execute(commands, queries)` — Primary tool. Runs all commands, auto-indexes output, returns search results. ONE call replaces 30+ individual calls.
2. **FOLLOW-UP**: `context-mode_ctx_search(queries: ["q1", "q2", ...])` — Query indexed content. Pass ALL questions as array in ONE call.
3. **PROCESSING**: `context-mode_ctx_execute(language, code)` | `context-mode_ctx_execute_file(path, language, code)` — Sandbox execution. Only stdout enters context.
4. **WEB**: `context-mode_ctx_fetch_and_index(url, source)` then `context-mode_ctx_search(queries)` — Fetch, chunk, index, query. Raw HTML never enters context.
5. **INDEX**: `context-mode_ctx_index(content, source)` — Store content in FTS5 knowledge base for later search.

## Output constraints

- Keep responses under 500 words.
- Write artifacts (code, configs, PRDs) to FILES — never return them as inline text. Return only: file path + 1-line description.
- When indexing content, use descriptive source labels so others can `search(source: "label")` later.

## ctx commands

| Command | Action |
|---------|--------|
| `ctx stats` | Call the `stats` MCP tool and display the full output verbatim |
| `ctx doctor` | Call the `doctor` MCP tool, run the returned shell command, display as checklist |
| `ctx upgrade` | Call the `upgrade` MCP tool, run the returned shell command, display as checklist |

<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **harvester** (940 symbols, 1919 relationships, 51 execution flows). Use the GitNexus MCP tools to understand code, assess impact, and navigate safely.

> If any GitNexus tool warns the index is stale, run `npx gitnexus analyze` in terminal first.

## Always Do

- **MUST run impact analysis before editing any symbol.** Before modifying a function, class, or method, run `gitnexus_impact({target: "symbolName", direction: "upstream"})` and report the blast radius (direct callers, affected processes, risk level) to the user.
- **MUST run `gitnexus_detect_changes()` before committing** to verify your changes only affect expected symbols and execution flows.
- **MUST warn the user** if impact analysis returns HIGH or CRITICAL risk before proceeding with edits.
- When exploring unfamiliar code, use `gitnexus_query({query: "concept"})` to find execution flows instead of grepping. It returns process-grouped results ranked by relevance.
- When you need full context on a specific symbol — callers, callees, which execution flows it participates in — use `gitnexus_context({name: "symbolName"})`.

## Never Do

- NEVER edit a function, class, or method without first running `gitnexus_impact` on it.
- NEVER ignore HIGH or CRITICAL risk warnings from impact analysis.
- NEVER rename symbols with find-and-replace — use `gitnexus_rename` which understands the call graph.
- NEVER commit changes without running `gitnexus_detect_changes()` to check affected scope.

## Resources

| Resource | Use for |
|----------|---------|
| `gitnexus://repo/harvester/context` | Codebase overview, check index freshness |
| `gitnexus://repo/harvester/clusters` | All functional areas |
| `gitnexus://repo/harvester/processes` | All execution flows |
| `gitnexus://repo/harvester/process/{name}` | Step-by-step execution trace |

## CLI

| Task | Read this skill file |
|------|---------------------|
| Understand architecture / "How does X work?" | `.claude/skills/gitnexus/gitnexus-exploring/SKILL.md` |
| Blast radius / "What breaks if I change X?" | `.claude/skills/gitnexus/gitnexus-impact-analysis/SKILL.md` |
| Trace bugs / "Why is X failing?" | `.claude/skills/gitnexus/gitnexus-debugging/SKILL.md` |
| Rename / extract / split / refactor | `.claude/skills/gitnexus/gitnexus-refactoring/SKILL.md` |
| Tools, resources, schema reference | `.claude/skills/gitnexus/gitnexus-guide/SKILL.md` |
| Index, status, clean, wiki CLI commands | `.claude/skills/gitnexus/gitnexus-cli/SKILL.md` |

<!-- gitnexus:end -->
