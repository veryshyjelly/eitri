# Eitri contributor instructions

## Scope and repository rules

- Eitri is a Go terminal companion for competitive programming: receive problems from Competitive Companion, prepare solution files, link a personal snippet library, and judge sample tests.
- Keep changes focused on the user's request. This onboarding adds project instructions and tool configuration; it does not change application behavior.
- **Do not create other Markdown files.** Keep project analysis and contributor instructions in this root `AGENTS.md`. Do not generate OpenSpec proposals, designs, tasks, specs, Serena Markdown memories, or additional instruction files unless the user explicitly changes this restriction. Leave the existing `README.md` alone unless asked to edit it.
- **Authorized exception:** the user requested the agreed website plan be added to OpenSpec for implementation. Proposal, design, tasks, and capability specs under `openspec/changes/add-eitri-onboarding-website/` are authorized. Maintain those artifacts as implementation progresses; this does not authorize unrelated Markdown files. The user subsequently authorized Cloudflare publication at eitri.veryshyjelly.in.
- Read working-tree source before editing. Treat `go.mod`, actual CLI flags, YAML structs, and key bindings as authoritative when the README differs.
- Preserve user changes. Do not upgrade dependencies, reformat unrelated files, or introduce a new build system as part of routine work.

## Required tools

### Serena: code navigation

- Call Serena's `initial_instructions` before code work and use its returned session ID. The project is `eitri`; its language server is Go.
- Start with `lsp.get_symbols_overview`, then narrow with `lsp.find_symbol`. Read bodies when needed and use `lsp.find_referencing_symbols` to trace callers before changing contracts.
- Prefer reference-aware rename and symbol edits where appropriate; use ordinary file edits for small changes and configuration.
- Do not write onboarding memories while the Markdown restriction is active. `.serena/project.yml` contains project configuration; local overrides and cache are ignored by `.serena/.gitignore`.

### Context7: current documentation

- Use Context7 whenever work depends on a library, framework, SDK, API, CLI, or cloud service: syntax, configuration, migration, platform ports, setup, or library-specific debugging. Use it even for familiar APIs.
- First call `resolve-library-id` with the library name and the actual documentation question, unless given an exact `/org/project` ID.
- Select by exact name, relevance, source reputation, snippet coverage, and benchmark score. Use a version-specific ID when available for the requested version; retry an unsuitable match with a better name or query.
- Call `query-docs` with that ID and a specific question scoped to one concept. Split unrelated topics into separate queries; combine only when investigating their interaction. Base library guidance on the fetched docs.
- If `query-code` is available and docs cannot establish an exact signature, default, option, error origin, or internal behavior, use a source-enabled repository ID. Start with `grep` (fixed strings for errors) or `definition`, then read the returned file range. Pass a git tag for versioned source; normally use 1–4 calls, stop at 6, and retry indexing at most twice. Cite source file and line when relying on it.
- General programming concepts and source-only refactoring or debugging that does not depend on an external API do not require Context7.
- Bubble Tea v2 documentation was resolved as `/charmbracelet/bubbletea/v2.0.0`; the project pins v2.0.7, so verify patch-specific behavior against the installed dependency. OpenSpec documentation is `/fission-ai/openspec`.

### Context-mode and search

- Use `rg`/`rg --files` for small known targets. Use context-mode for large files, logs, searches, and test output; print conclusions and relevant excerpts instead of dumping entire datasets.
- Use `ctx_resume` after resuming work, and `ctx_index`/`ctx_search` for material that will be queried repeatedly.
- CodeGraphContext and jcodemunch are opt-in. Refresh their indexes after edits, read working-tree source before changing it, and check locking support before sharing an embedded graph database across processes.

### OpenSpec: exploration within the file constraint

- OpenSpec is initialized with `schema: spec-driven` in `openspec/config.yaml`. Setup used `openspec init --tools none --no-animation` to avoid generated tool skills and Markdown files.
- Check existing context before planning: `openspec list --json`, `openspec list --specs --json`, and `openspec context --json`. `openspec doctor --json` checks workspace relationships.
- At onboarding there were no specs or changes. Do not imply that application behavior has formal OpenSpec coverage.
- Use OpenSpec's exploration approach to investigate current behavior and clarify intended changes. Exploration does not require generating artifacts; record useful findings in `AGENTS.md` under the current restriction.
- Do not run artifact-generating workflows such as `openspec new change`, proposal generation, or integration refreshes that create other Markdown files while this restriction remains active.
- The website change named above is an explicit exception to that workflow restriction. Read its proposal, design, specs, and tasks before implementing the website; leave tasks unchecked until their implementation and verification are complete.
- If the user later authorizes standard spec artifacts, consult the installed CLI help and current docs, follow the configured schema, and validate actual artifacts with `openspec validate --all --strict --no-interactive`. An empty workspace is not evidence that specifications passed validation.

## Stack and source map

The module is `github.com/veryshyjelly/eitri`. `go.mod` requires **Go 1.25.11**; the README's Go 1.22 minimum is outdated. This is one Go executable with local packages, no database, and no frontend build pipeline.

| Location | Responsibility |
| --- | --- |
| `main.go` | Kong CLI, logging, first-run config selection, startup validation, watcher and HTTP server setup, Bubble Tea program |
| `config/config.go` | YAML structures and decoding with `github.com/goccy/go-yaml` |
| `config/validate.go` | Aggregate configuration diagnostics, regex/template checks, required fields, and filesystem checks |
| `config/validate_test.go` | Existing automated tests; valid/invalid configurations and diagnostic formatting |
| `core/types.go` | `App`, problem metadata, incoming JSON, testcase results, and verdicts |
| `core/files.go` | Filename routing, boilerplate creation, editor execution, clipboard/compiled output, library discovery |
| `core/linker.go` | Snippet dependency graph, DFS ordering/cycle detection, final solution rendering |
| `core/utils.go` | `@head`/`@code` extraction and fatal error handling/terminal restoration |
| `core/judge.go` | Temporary compilation directory, concurrent testcase execution, verdict aggregation |
| `server/server.go` | Problem intake and one-shot submission JSON endpoints |
| `watcher/watcher.go` | Active solution's parent-directory watch and 200 ms write debounce |
| `tui/model.go` | Main model, event dispatch, problem selection, pane content and geometry |
| `tui/home.go`, `tui/styles.go` | Lip Gloss rendering, result display, theme, mouse hit rectangles |
| `tui/splash.go`, `tui/help.go`, `tui/add.go` | Waiting screen, help screen, in-memory manual testcase form |
| `tui/keys.go` | Authoritative key bindings and help groups |
| `tmpl/funcs.go` | Shared Go-template string/case helpers |
| `templates/` | Embedded C++, Rust, Python, and OCaml starter YAML configurations |
| `memory/` | OS-specific peak RSS reporting; Windows stub |

The UI uses `charm.land/bubbletea/v2`, `bubbles/v2`, `huh/v2`, `lipgloss/v2`, and `log/v2`. Other central dependencies are Kong, fsnotify, google/shlex, atotto/clipboard, samber/lo, and ettle/strcase. Keep the v2 Charm import paths and APIs; `View()` returns `tea.View`, configured with `tea.NewView`, `AltScreen`, and `MouseMode`.

## Execution and data flow

1. `main` parses options, handles `--version`, and logs to `filepath.Join(os.TempDir(), "eitri.log")`. `--debug` enables debug logging.
2. Read `--config` (default `./eitri.yml`). If missing, show a Huh language picker, copy an embedded YAML starter to that path, then exit. The user must create the referenced source/library paths and edit the starter before restarting.
3. `config.Validate` runs before watcher/server/TUI startup. It accumulates `Issue{Path, Message, Hint}` values in `ValidationError`; startup prints failures to stderr and exits 1.
4. Start a Bubble Tea splash model, fsnotify loop, and HTTP listener on **`127.0.0.1:6174`**. `/` decodes Competitive Companion JSON into `core.Info`; invalid JSON returns 400, valid JSON is sent through `Program.Send` and returns 200.
5. `Model.setProblem` replaces the active metadata/tests. Incoming `tests[].output` becomes `Testcase.Answer`; `Output` is reserved for actual execution stdout. The app tracks one active problem and stores tests in memory.
6. On problem intake, optionally create the solution (`create_file`), send its name through `fileChan`, and open the configured editor if set. This does not immediately run tests.
7. The watcher observes the active file's parent, filters by basename and `fsnotify.Write`, and debounces writes for 200 ms. The TUI marks the file changed and schedules `App.Run` when `run_on_save` is true; the run key also schedules it.
8. `App.Run` returns `[]core.Testcase` as a Bubble Tea message. The model replaces displayed results and derives the overall verdict.
9. Copy sets `server.Solution` and snapshots `server.App`, then schedules clipboard work. `/submit` consumes the flag and returns `empty`, `problemName`, `url`, `sourceCode`, and `fileName`; without a pending copy it returns `empty: true`.

HTTP intake and watcher goroutines communicate through Bubble Tea messages. `App.Run` and `CopyFile` have `func() tea.Msg` command-compatible signatures; `OpenEditor` uses `tea.ExecProcess`. Preserve this event flow when changing asynchronous work.

## Configuration and path contracts

- CLI options are `-c/--config`, `-C/--root`, `-l/--vertical`, `-v/--version`, and `-d/--debug`, plus help. `--root` is an existing directory and does not change the process working directory or automatically relocate `--config`.
- YAML sections are `author`, `editor`, `filename.rules`, `template`, `code`, `lib`, `compiler`, `create_file`, `run_on_save`, and optional `compiled_file`. The starter's `version` key has no corresponding `Config` field.
- Workspace source/template/library locations are joined with `App.Root`. Compiler source and compiled output are joined with the temporary compilation directory. Avoid conflating these bases.
- `GetFileName` lowercases the URL hostname and strips `www.`, then chooses the **first** rule with an exact `site` match. Its regex runs on the complete URL; filename templates receive `.Captures` (index 0 is the full match) and a `.Title` with single/double quotes removed. An unmatched site exits the process.
- `CreateFile` skips a file it can already read and otherwise creates it directly; it does **not** create parent directories. Its template receives `.Author`, `.Time`, `.Url`, `.Title`, `.Filename` (basename without extension), and `.Code` (the base template file's content).
- Editor templates receive the root-joined `.Filename`; the rendered command is shlex-split and run with `cmd.Dir = App.Root`. Editor and run command strings do not execute through a shell, so shell operators are not interpreted.
- `compiler.compile` is an executable with a separate `compiler.args` list. Empty compile skips compilation. `compiler.run` is shlex-split. `compiler.source` names the file containing the linked solution in the temporary directory.
- `compiled_file` causes copy/submission to compile again and read that output. Empty content, compilation failure, or unreadable output falls back to the linked source. Keep clipboard and HTTP submission behavior aligned.
- Validation requires filename rules, a template source that exists, compiler source/run, compile arguments when a compiler is set, and existing include paths. Filename templates are executed with dummy captures; optional templates are parsed for syntax. It does not prove every actual URL matches, validate every rendered library regex, or check compiler executables are installed.
- Use `tmpl.FuncMap` consistently in validation and runtime rendering. `stripPrefix` removes everything through the first delimiter; `removeLines` drops leading lines.

## Snippet linking contracts

- `lib.include` maps names to individual files or directories. Individual files keep their configured name. Directories load all immediate entries that are not directories, keyed by filename stem; discovery is nonrecursive and does not filter by the solution's extension.
- Render `lib.regex` with `.Name` for each candidate, match against library contents to build dependencies, and match against the solution to select roots.
- DFS includes only reachable libraries and orders dependencies before dependants. A gray-node revisit reports a cycle and exits. Independent-library ordering is not guaranteed because maps are traversed.
- `@head begin` / `@head end` delimit header blocks; `@code begin` / `@code end` delimit code blocks. Without markers, code defaults to the entire file and headers default to empty. Malformed/empty marker behavior comes from `extractBlock`; inspect it before changing extraction.
- Header lines from linked files and the solution are trimmed and deduplicated. `code.modifier` receives `.Author`, `.Url`, `.Time`, `.LibFiles` (name/content), `.Header`, and `.Code`.
- Compilation, ordinary clipboard copy, and ordinary HTTP submission all share `GetSolution`. Changes there affect all three paths; do not edit the user's library files while composing output.

## Judging and platform behavior

- Compilation creates an `os.MkdirTemp` directory, writes the linked solution under `compiler.source`, runs the configured compiler there, and captures stderr. Compile failure sets all test verdicts to compilation error.
- Testcases run concurrently in that same directory with separate processes/stdin/stdout/stderr. Successful runs and successful compiled-output reads remove the temporary directory; compilation failures can leave it behind.
- `TimeLimit` is milliseconds; an after-function timer kills the direct child process on timeout. `Testcase.Time` is elapsed seconds and `Memory` is KiB. `MemoryLimit` is stored but not enforced. Compilation has no timeout.
- Verdict order per test is timeout, process error, trimmed stdout equality, then wrong answer. Comparison uses `strings.TrimSpace` on both output and answer, preserving internal whitespace.
- Overall verdict precedence is compilation error, runtime error, time limit exceeded, wrong answer, accepted. `GetFinalStatus` returns accepted for an empty slice.
- The Unix memory implementation reads `syscall.Rusage.Maxrss`: Linux reports KiB; macOS bytes are divided by 1024. Windows returns `(0, false)` and omits memory in the status display.
- The temporary working directory is process workspace separation, not OS-level isolation. There is no memory enforcement, process-tree containment, or interactive judge; the incoming `Interactive` field is not used.

## UI behavior and edges to consider

- Home has four pane modes: input/answer, input/output, input/error, and answer/output. Horizontal means side by side; vertical means stacked. Resize and mouse handling depend on matching viewport sizes and `Rect` geometry.
- Actual bindings include `q`/`Esc`/`Ctrl+D` to quit, `r` to run, `f` to create, `c`/`Ctrl+C` to copy, `a`/`Ctrl+N` to add a testcase, `1`–`4` or `j/k/l/;` for panes, Tab/arrows for cases, `?` for help, and `Ctrl+H`/`Ctrl+V` for orientation. The README's Ctrl+C-to-quit description differs from `DefaultKeyMap`.
- Manual tests are held only in memory. Help and add-case views are separate models; account for message forwarding when editing modal behavior.
- Results currently arrive in completion order. `Run` goroutines share an outer `err`, and HTTP submission uses unsynchronized package globals; concurrency changes need focused review and tests.
- Exercise zero-sample problems, rapid problem switches, overlapping runs, modal views during execution, small terminal dimensions, and editor saves implemented with rename. Existing tests do not cover these paths; the watcher handles write events only.
- The UI already shows per-selected-test time and, on non-Windows systems, memory for overall AC/WA states. Do not assume every README roadmap item is unimplemented without checking source.

## Development and verification

Run from the repository root:

```sh
go test ./...
go vet ./...
go build -o /tmp/eitri .
/tmp/eitri --help
/tmp/eitri --version
```

- During onboarding, `go test ./...`, `go vet ./...`, a build to `/tmp/eitri-inspect`, and that binary's help/version checks passed. Existing automated tests are limited to `config/validate_test.go`; build/test success does not establish end-to-end watcher, browser, compiler, clipboard, or terminal behavior.
- For Go changes, format only edited files with `gofmt -w`, run focused tests and `go test ./...`, and use `go vet ./...`/build as appropriate. Add behavioral regression tests for substantive fixes rather than tests that duplicate implementation.
- Use `t.TempDir()` for filesystem fixtures, following the configuration tests. For execution changes, use controlled small subprocess fixtures and explicit limits. For concurrent code, use the race detector on tests that actually exercise that code.
- Full interactive verification needs a real terminal, a configured external editor/compiler/interpreter, existing workspace directories, and Competitive Companion or equivalent local JSON input. Use a disposable workspace, not the repository as a contest workspace.
- Consult `go.mod` for exact versions. No Makefile, CI workflow, or release automation is present; do not invent repository-specific lint or release commands.
- Preserve platform build tags in `memory/`, keep diagnostic paths/hints useful, and avoid writing logs to the TUI's stdout. Many existing fatal paths exit the process through `Unwrap` or `os.Exit`; account for that when designing tests or changing error handling.

Documentation references used for tool/API guidance: [OpenSpec CLI](https://github.com/fission-ai/openspec/blob/main/docs/cli.md), [OpenSpec exploration](https://github.com/fission-ai/openspec/blob/main/docs/explore.md), and [Bubble Tea v2 migration guide](https://github.com/charmbracelet/bubbletea/blob/v2.0.0/UPGRADE_GUIDE_V2.md). Repository behavior above is derived from the working-tree source.

## Website development

The local onboarding website is in `website/`. It uses a dependency-free Node generator and plain HTML/CSS/JavaScript; Go application behavior is unchanged.

- Build: `node website/build.mjs` (Node 20+). Output: `website/dist/` (ignored).
- Preview: `npm --prefix website run dev`, then open `http://127.0.0.1:4173`. Stop with Ctrl+C. The preview binds loopback only.
- Verify content and examples: `npm --prefix website test`. This validates twelve starter configurations plus twelve adapted recipe/path manifests with Eitri's real validator, verifies and compiles three handbook transformations, and compiles/runs four starter programs using installed toolchains. Recipe libraries use path placeholders during validation; this is not a full-project compilation check. Run from any folder via `npm --prefix /absolute/path/to/eitri/website test`.
- Source content: `website/content.mjs`; generator: `website/build.mjs`; browser interactions: `website/app.js`; URL selection: `website/selection.js`.
- Keep runtime Go changes out of website work. `website/tests/validate.go` is a standalone verification helper, not application code. Go tests/vet can still run from the repository root.
- Preserve original tutorial GIFs. The website uses an optimized MP4, JPEG poster, and WebVTT captions. The demonstration recording uses an AtCoder problem and an advanced setup; the beginner guide uses Codeforces Watermelon.
- Production publication at eitri.veryshyjelly.in is authorized. The generator sets canonical/social URLs for that hostname. Native Windows/Linux end-to-end verification is outstanding; do not describe it as completed.
- Optional macOS integration smoke: `python3 website/tests/smoke.py` builds the CLI into `/tmp`, starts a disposable PTY/workspace, sends synthetic Competitive Companion JSON, and verifies save-triggered AC. Requires free loopback port 6174; it does not click the extension, submit to a judge, or test the desktop clipboard.

- Website typography uses self-hosted Monaspace Neon across headings, body text, controls, captions, and code. Keep its variable WOFF2 and SIL license in `website/assets/fonts/`; do not introduce a different UI typeface.

### Expanded website documentation

- Routes: `/`, `/get-started/`, `/setups/` (Languages & toolchains), `/handbook/`, and `/reference/`. Keep direct links and section anchors working without client-side routing.
- `website/docs.mjs` explains Eitri's execution, source transformation, build and integration model. `website/reference.mjs` describes every current YAML field, template context and helper. `website/transformations.mjs` loads shared explanatory examples from test fixtures.
- The user clarified that cp_x and local scripts are research guidance only. Do not publish personal workspaces, source-branch provenance, clone commands, helper scripts or their configuration catalog. Explain the generic tool with focused examples instead.
- Internal fixtures under `website/tests/fixtures/` exercise research-derived contracts. Do not copy them into generated output. Builds clear the ignored dist directory to remove obsolete published assets.
- `website/tests/documentation.test.mjs` checks links, assets, anchors, current schema/helper coverage and the exclusion of research-workspace material from the site. Preserve Monaspace throughout and captions off by default.

### Cloudflare publication

- `website/wrangler.jsonc` defines a static-assets-only Worker named eitri-website on the authorized custom domain eitri.veryshyjelly.in. It uses the existing Cloudflare account and no runtime bindings. The workers.dev hostname and preview URLs are disabled.
- Deploy with `npm --prefix website run deploy` after the website checks. The script builds before invoking pinned Wrangler 4.162.0; credentials stay in Wrangler’s user-level login store, never in the repository.
- The generator emits canonical/absolute social URLs, robots.txt, sitemap.xml and a custom 404 page. dist/.assetsignore excludes the starter test-fixture JSON from uploads. Keep internal research fixtures out of generated assets.
- Verify HTTPS, all five routes, media/font assets, trailing-slash redirects, custom 404 behavior and absence of test fixtures after publishing. The hosting source of truth is Cloudflare, not Sites hosting.
