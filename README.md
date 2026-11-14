# ⚒️ Eitri

**Pick a problem. Write your solution. Save to test.**

Eitri is a terminal companion for competitive programming. Send a problem from your browser, edit your solution in your preferred editor, and see sample-test results whenever you save. When you are ready, copy the prepared source for submission.

Your language, editor, and toolchain are configurable. Eitri runs the commands and follows the source-generation rules you provide; its bundled language starters are examples, not a fixed support list.

**[Get started](https://eitri.veryshyjelly.in/get-started/)** · [Website](https://eitri.veryshyjelly.in/) · [Documentation](https://eitri.veryshyjelly.in/handbook/) · [Releases](https://github.com/veryshyjelly/eitri/releases)

[![License: GPL-2.0](https://img.shields.io/badge/License-GPL--2.0-yellow.svg)](LICENSE)

[![Eitri showing sample-test results alongside Zed](use_with_zed.png)](https://eitri.veryshyjelly.in/#demo)

[Watch the browser → editor → test workflow](https://eitri.veryshyjelly.in/#demo).

## What Eitri handles

- **Receive problems and samples** through the [Competitive Companion](https://github.com/jmerle/competitive-companion) browser extension.
- **Prepare solution files** using your naming rules and starter templates, with optional editor launch.
- **Run samples on save**, showing input, expected output, actual output, and errors in the terminal.
- **Prepare submission source**, optionally combining selected snippets or copying a text artifact produced by your build.

Use a compiler, an interpreter, or a project build command. Keep source transformations and snippet linking optional. Your installed toolchain determines what runs locally; the online judge determines what it accepts.

## Install

### macOS and Linux · Homebrew

```sh
brew tap veryshyjelly/eitri https://github.com/veryshyjelly/eitri.git
brew install veryshyjelly/eitri/eitri
```

### Windows · Scoop

```powershell
scoop bucket add eitri https://github.com/veryshyjelly/eitri.git
scoop install eitri/eitri
```

Windows x64 and ARM64 packages are available.

### With Go

Requires [Go](https://go.dev/dl/) **1.25.11 or newer**:

```sh
go install github.com/veryshyjelly/eitri@latest
```

You will also need your language's compiler or runtime and the Competitive Companion browser extension. The [first-run guide](https://eitri.veryshyjelly.in/get-started/) covers these prerequisites and provides a complete workspace and configuration.

## Your first problem

1. Follow the [first-run guide](https://eitri.veryshyjelly.in/get-started/) to create a workspace and `eitri.yml` for your language and operating system.
2. Run `eitri` from that workspace and keep it open.
3. Open a supported problem page and click Competitive Companion to send the problem and samples to Eitri.
4. Edit the solution file. With `run_on_save` enabled, saving runs the samples; press `r` to run them manually.
5. Press `c` to copy the prepared solution, then paste it into the judge's submission page.

If you run `eitri` without a configuration, it offers to write a starter and then exits. Adjust the configuration and create its referenced files before starting Eitri again. Automatic file creation and editor launch depend on your configuration.

For optional browser submission automation, see [cp-submit](https://github.com/veryshyjelly/cp-submit) for CSES, AtCoder, and Codeforces. Read the [integration notes](https://eitri.veryshyjelly.in/handbook/#submission) before enabling it, including compatibility limits and the manual fallback.

## Go deeper

| You want to… | Read |
| --- | --- |
| Complete your first sample test | [Get started](https://eitri.veryshyjelly.in/get-started/) |
| Use another language, runtime, or project build | [Languages & toolchains](https://eitri.veryshyjelly.in/setups/) |
| Understand file creation, source transformation, and snippet linking | [How Eitri works](https://eitri.veryshyjelly.in/handbook/) |
| Look up a YAML field, template value, or helper | [Configuration reference](https://eitri.veryshyjelly.in/reference/) |
| Diagnose intake, path, build, or submission problems | [Troubleshooting](https://eitri.veryshyjelly.in/handbook/#troubleshooting) |

## Terminal essentials

Run `eitri --help` for all options. Use `--config <file>` to select a configuration, `--root <dir>` to select the workspace, and `--vertical` for stacked panes. Selecting a config file does not change the workspace root.

| Key | Action |
| --- | --- |
| `r` | Run samples |
| `c` | Copy the prepared solution |
| `a` | Add a custom test case |
| `f` | Create the solution file from its template |
| `Tab` / `Shift+Tab` | Next / previous test case |
| `1`–`4` | Switch input, answer, output, and error views |
| `?` | Show all keybindings |
| `q` / `Esc` | Quit |

Sample tests are local feedback, not the judge's final verdict. Comparison trims whitespace at the ends of the complete output; internal whitespace remains significant. Programs run with your user permissions, and temporary build directories are not a security sandbox. Memory usage is informational, not an enforced limit.

## Help and contributions

For a bug report or feature request, [open an issue](https://github.com/veryshyjelly/eitri/issues). Include your Eitri version (`eitri --version`), operating system, relevant configuration, and steps to reproduce the problem. Debug logging is available with `eitri --debug`; logs are written to `eitri.log` in your operating system's temporary directory.

To work on the CLI, install the Go version required by [go.mod](go.mod), then run:

```sh
go test ./...
go vet ./...
go test -race ./...
go build ./...
```

The documentation website lives in `website/` and requires Node.js 20 or newer:

```sh
npm --prefix website test
npm --prefix website run dev
```

The preview runs at `http://127.0.0.1:4173`. Website tests also use Go; language execution checks depend on the corresponding local toolchains.

## License

[GNU General Public License v2.0](LICENSE).
