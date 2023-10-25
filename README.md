# gochange

> Run commands on file change using glob patterns, heavily inspired by the excellent [onchange][onchange], but written in Go.

## Install

```shell
go install github.com/danawoodman/gochange
```

## Usage

```shell
# Run `go run ./cmd/myapp` when any .go or .html files change.
# `-i` runs the command once on load without any event
# `-k`` kills running processes between changes
# The command you want to run is followed by the `--` separator
gochange -i -k '**/*.go' 'templates/**/*.html' -- go run ./cmd/myapp
```

## Options

```shell
$ gochange
Runs a command when file changes are detected

Usage:
  gochange [flags] [paths] -- [command]

Flags:
  -a, --add               Execute command for initially added paths
  -d, --delay int         Delay between process changes
  -e, --exclude strings   Exclude matching paths
  -h, --help              Help for gochange
  -i, --initial           Execute command once on load without any event
  -k, --kill              Kill running processes between changes
  -v, --verbose           Enable verbose logging
```

## Notes and Limitations

Currently, gochange only supports a subset of the onchange commands, but I'm open to adding more. Please open an issue if you have a feature request.

This is a very new project and hasn't been tested really anywhere outside of my machine (macOS), if you run into any issues, please open an issue!

No test suite as of yet, but I aspire to add one 😇.

## Motivations

Mostly, this project was an excuse to play more with Go, but also I wanted a more portable version of onchange.

I also couldn't find the tool I wanted in the Go (or broader) ecosystem that was a portable binary. I tried out [air][air], [gow][gow], [Task][task] and others but none of them really fit my needs. I loved onchange but the combo of requiring Node, not being maintained anymore, and not being a portable binary was a deal breaker for me (well that and I just wanted to try and make it myself in Go 😅).

## Development

PRs welcome!

Code lives in `internal` and the CLI is in `cmd/gochange`, using Cobra to parse command line arguments.

```shell
# Build the CLI
make build

# Install the CLI locally
make install

# Run the CLI in watch mode using gochange :)
# Pass the arguments to gochange using the ARGS variable.
# Make sure to run `make build` first!
make dev ARGS="-i -k 'some/path/*.html' -- echo 'changed'" # or just `make`
```

## License

MIT

## Credits

Written by [Dana Woodman](https://danawoodman.com) with heavy inspiration from [onchange][onchange].

[onchange]: https://github.com/Qard/onchange
[air]: https://github.com/cosmtrek/air
[gow]: https://github.com/mitranim/gow
[task]: https://github.com/go-task/task
