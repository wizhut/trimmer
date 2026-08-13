# trimmer

A tiny command-line filter: read stdin, write it back to stdout with the leading
and trailing whitespace gone.

```bash
$ printf '   hello world   \n\n' | trimmer
hello world
```

By default the whole input is treated as a single value, so only the very start
and the very end are trimmed — inner blank lines and indentation are left alone.

## Install

```bash
go install github.com/wizhut/trimmer@latest
```

Or build from source:

```bash
make build
```

The binary lands in `bin/trimmer`.

### Cross-compiling

```bash
make all        # every platform below
make mac        # darwin/amd64 + darwin/arm64
make linux      # linux/amd64 + linux/arm64
make windows    # windows/amd64 + windows/arm64
```

Individual targets are also available: `mac-intel`, `mac-arm`, `linux-amd64`,
`linux-arm64`, `windows-amd64`, `windows-arm64`. Binaries are written to
`bin/trimmer-<goos>-<goarch>` (`.exe` on Windows), built with `CGO_ENABLED=0`
and `-trimpath -ldflags "-s -w"` so they are static and stripped.

`make release` then packages each build together with this README into `dist/`
— `.tar.gz` for macOS/Linux, `.zip` for Windows, version-stamped from the
`version` const in `main.go`:

```
dist/trimmer-1.0.0-darwin-arm64.tar.gz
dist/trimmer-1.0.0-linux-amd64.tar.gz
dist/trimmer-1.0.0-windows-amd64.zip
...
```

## Usage

```
trimmer [flags]

Flags:
  -l, --lines        Trim each line individually instead of the whole input
      --left         Trim leading whitespace only
      --right        Trim trailing whitespace only
  -n, --no-newline   Do not terminate the output with a newline
      --version      Show version information
  -h, --help         Help for trimmer
```

### Examples

Strip the surrounding whitespace of a command's output — useful when a value is
about to be interpolated somewhere:

```bash
VERSION=$(cat VERSION | trimmer -n)
```

Trim every line of a file, keeping the line structure intact:

```bash
trimmer --lines < notes.txt
```

Left-strip indentation only, leaving trailing spaces untouched:

```bash
trimmer --lines --left < indented.txt
```

## Behaviour

| Input | Command | Output |
|---|---|---|
| `"  hi  "` | `trimmer` | `"hi\n"` |
| `"  hi  "` | `trimmer -n` | `"hi"` |
| `"\n\n  a\n  b  \n\n"` | `trimmer` | `"a\n  b\n"` |
| `"\n\n  a\n  b  \n\n"` | `trimmer -l` | `"\n\na\nb\n\n"` |
| `"  hi  "` | `trimmer --left` | `"hi  \n"` |
| `"  hi  "` | `trimmer --right` | `"  hi\n"` |

Notes:

- Whitespace is Unicode-aware — anything matching `unicode.IsSpace`, so tabs and
  carriage returns count, and so do runes like NBSP (U+00A0) and U+2028.
- `--lines` preserves the line count: blank lines stay as blank lines.
- `--lines` streams, so it handles inputs larger than memory. The default
  whole-input mode has to buffer, since the trailing whitespace run is only
  known once the input ends.
- CRLF input is normalised to LF in `--lines` mode (the `\r` is whitespace).
- Running `trimmer` with no piped stdin prints the usage text rather than
  hanging on the terminal.

## Development

```bash
make test    # run the test suite
make vet     # go vet
make build   # build to bin/
```

## Licence

Provided by [wizhut.tech](https://wizhut.tech). Free to use for any purpose.
