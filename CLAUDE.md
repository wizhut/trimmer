# CLAUDE.md

Guidance for Claude Code (claude.ai/code) when working in this repository.

## What this is

`trimmer` is a single-purpose Go CLI: it reads stdin and writes it back with the
leading/trailing whitespace removed. Everything lives in one package at the repo
root — [main.go](main.go) and [main_test.go](main_test.go). There is no internal
package layout, and adding one would be overkill for the scope of this tool.

Module path: `github.com/wizhut/trimmer` (matches the remote, so
`go install github.com/wizhut/trimmer@latest` works). Go 1.26. Dependencies are
`github.com/spf13/cobra` — matching the CLI convention used by the `devbin`
tools in this workspace — and `golang.org/x/term` for the TTY check.

## Commands

```bash
make build   # -> bin/trimmer
make test    # go test with a summarised pass/fail report
make vet     # go vet
make all     # cross-compile darwin/linux/windows (amd64+arm64 each) into bin/
make mac     # or: make linux / make windows — one OS, both arches
make release # package the cross-builds into dist/
make publish # release + upload the archives to the public GCS bucket
```

Cross-compilation runs through the `cross` macro in the
[Makefile](Makefile) — `$(call cross,<goos>,<goarch>,<suffix>)`. Add a platform
by adding one target plus an entry in `UNIX_PLATFORMS` / `WIN_PLATFORMS`; the
`release` loop picks it up from those lists. `VERSION` is scraped out of
`main.go` with `sed`, so release archives stay in step with `--version`.

## Design decisions worth preserving

- **Two modes.** Default treats the whole input as one value (must buffer, since
  the trailing whitespace run is only known at EOF). `--lines` streams line by
  line via `bufio.Reader.ReadString` — deliberately *not* `bufio.Scanner`, which
  would cap lines at 64 KiB. `TestTrimLinesLongLine` guards that choice.
- **Lazy newline in `--lines` mode.** `trimLines` defers writing a line's
  terminator until the next line arrives, so `--no-newline` can drop only the
  final one instead of joining every line together.
- **Flag-backed globals.** `perLine` / `leftOnly` / `rightOnly` / `noNewline` are
  package-level and read by `run` and `trim`. Tests call `reset()` first — keep
  doing that in any new test, or state leaks between cases.
- **`--left` and `--right` together** mean both sides, same as passing neither.
- **TTY guard.** With no piped stdin, `trimmer` prints usage rather than blocking
  on the terminal. `isTerminal` uses `golang.org/x/term`, *not* a
  `os.ModeCharDevice` check — `/dev/null` is a character device too, and
  `trimmer < /dev/null` should trim an empty input rather than print usage.
- **Company notice lives in the usage template,** not in `Long`. Appending it via
  `SetUsageTemplate` means it renders under the flags on every path — `--help`,
  the TTY guard's usage, and (printed separately) `--version` — instead of only
  on `--help`. Don't duplicate it back into `Long`.

## Distribution

Two things outside this repo depend on the archive names, so
`$(APP_NAME)-<goos>-<goarch>-$(VERSION).<ext>` is a contract, not a preference:

- **Public bucket.** `make publish` uploads `dist/` to
  `gs://wizhut-tools-downloads/trimmer/`, the same bucket json2jsonl and Contain
  use. Objects are world-readable via the bucket's `allUsers:objectViewer`
  binding — no per-object ACL step — and serve from
  `https://storage.googleapis.com/wizhut-tools-downloads/trimmer/<file>`.
- **Landing page.** `corporate-site/pages/trimmer.vue` builds those exact
  filenames from a `VERSION` const. Bumping the version here means bumping it
  there too, along with the `lastmod` in the site's `sitemap.urls` and a new
  [whats-new.vue](../../corporate-site/pages/whats-new.vue) entry — the
  release-coupling workflow in
  [corporate-site/CLAUDE.md](../../corporate-site/CLAUDE.md).

## Version

`version` is a const in [main.go](main.go). Bump it there when cutting a release
— the Makefile scrapes it for the `dist/` archive names, so it is the only place
to change in this repo.

## Licence

MIT, `Copyright (c) 2026 wizhut.tech` — same as the sibling `open-source/`
projects. [LICENSE](LICENSE) ships inside every release archive, and the licence
is named in `companyNotice` (shown on `--help` / `--version`) and the README.
