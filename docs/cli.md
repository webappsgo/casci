# CLI reference

The `casci` executable starts the server when invoked without an informational
or management flag. It is a single server binary; the repository does not yet
ship separate client or build-agent binaries.

```text
casci [flags]
```

## Information flags

| Flag | Behavior |
|------|----------|
| `--help`, `-h` | Print help |
| `--version`, `-v` | Print version and build metadata |
| `--status` | Query the configured local health endpoint; zero means healthy |

## Server flags

| Flag | Effect |
|------|--------|
| `--mode` | `production`, `development`, or `debug`; `prod`, `dev`, and `devel` are accepted aliases |
| `--config DIR` | Configuration directory |
| `--data DIR` | Data directory |
| `--cache DIR` | Cache directory |
| `--log DIR` | Log directory |
| `--backup DIR` | Backup directory |
| `--pid FILE` | Accepted, but no PID file lifecycle is implemented |
| `--address ADDR` | Listen address |
| `--port PORT` | Listen port, from 1 through 65535 |
| `--baseurl PATH` | Accepted and checked to start with `/`; path-prefix routing is not implemented |
| `--daemon` | Accepted; daemon detachment is not implemented |
| `--debug` | Enable debug profiling and detailed logging, regardless of application mode |
| `--color auto\|yes\|no` | Select color output behavior |
| `--lang CODE` | Accepted; CLI localization is not implemented |

Debug endpoint routing is not implemented yet. `--shell`, `--service`, `--maintenance`, and `--update` are accepted command
selectors but their operations are incomplete. They return an error for
unimplemented commands. Use `--help` for the flags actually parsed by the
current binary.

## Color output

`--color=yes` forces color, and `--color=no` disables it. In `auto` mode,
`NO_COLOR` or `TERM=dumb` disables color. These variables also disable emoji
output where applicable.

## Examples

```bash
casci --help
casci --version
casci --status
casci --mode development --debug
casci --address 127.0.0.1 --port 64580
```

For configuration paths and environment variables, see
[Configuration](configuration.md). For service scripts, see
[Installation](installation.md).
