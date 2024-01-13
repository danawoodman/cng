# 🔭 gochange

<div style="align: center">
	<img src="https://p198.p4.n0.cdn.getcloudapp.com/items/X6uXY0mB/3d7fe08d-4169-4837-afe5-765d8f6e73fd.png?v=b86f40978b79aaaec626ae521db89d58" alt="gochange logo" />
</div>

> Run commands on file change using glob patterns, heavily inspired by the excellent [onchange][onchange], but written in Go.

## Install

```shell
go install github.com/danawoodman/gochange
```

**Coming soon(?):** Downloads for macOS, Linux, and Windows...

## Usage

```shell
# Run `go run ./cmd/myapp` when any .go or .html files change.
# `-i` runs the command once on load without any event
# `-k`` kills running processes between changes
# The command you want to run is followed by the `--` separator:
gochange -i -k '**/*.go' 'templates/**/*.html' -- go run ./cmd/myapp

# Run tests when your source or tests change:
gochange 'app/**/*.tsx?' '**/*.test.ts' -- npm test

# Wait 500ms before running the command:
gochange -d 500 '*.md' -- echo "changed!"

# Ignore/exclude some paths:
gochange -e 'path/to/exclude/*.go' '**/*.go' -- echo "changed!"
```

## Features

- Watch for changes using global patterns like `'*.go'` or `'src/**/*.jsx?'` (using [doublestar][doublestar], which is a much more flexible option than Go's built in glob matching). Watching is done using the very fast [fsnotify][fsnotify] library.
- Run any command you want, like `go run ./cmd/myapp` or `npm test`
- Optionally kill running processes between changes, useful for when running web servers for example. Importantly, gochange kills all child processes as well, so your ports get properly freed between runs (avoids errors like `"bind: address already in use"`)
- Optionally run the task immediately or only run when a change is detected (default)
- Pass in a delay to wait between re-runs. If a change is detected in the delay window, the command will not be re-run. This is useful for when you're making a lot of changes at once and don't want to run the command for each change.
- Optionally exclude paths from triggering the command

## Options

```
$ gochange
Runs a command when file changes are detected

Usage:
  gochange [flags] [paths] -- [command]

Flags:
  -a, --add               Execute command for initially added paths
  -d, --delay int         Delay between process changes in milliseconds
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

I also couldn't find the tool I wanted in the Go (or broader) ecosystem that was a portable binary. I tried out [air][air], [gow][gow], [Task][task] and others but none of them really fit my needs (still great tools tho!). For me, air didn't work well when I tried it with `go run`. `gow` does work with `go run` but it's not generic enough to use outside of go projects. `Task` is a cool modern alternative to make but I also could get it working well with `go run` and killing my web server processes (and associated port binding).

I loved onchange but the combo of requiring Node, not being maintained anymore, and not being a portable binary was a deal breaker for me (well that and I just wanted to try and make it myself in Go 😅).

## Development

PRs welcome!

- We use Cobra to parse command line arguments

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
[doublestar]: https://github.com/bmatcuk/doublestar
[fsnotify]: https://github.com/fsnotify/fsnotify
