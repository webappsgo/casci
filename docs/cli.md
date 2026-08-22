# CLI Reference

The server binary is single-command: running it with no arguments starts the server. There are no subcommands, only flags.

```
casci [flags]
```

The binary can be renamed. Help, version, and error output always show the actual filename it was invoked as, while internal identifiers such as the User-Agent and the default paths stay tied to `casci`.

## Information

| Flag | Description |
|------|-------------|
| `-h`, `--help` | Show help. `--help` after any command shows that command's help |
| `-v`, `--version` | Show version |
| `--status` | Show server status and health |

`--status` queries the running server's health endpoint and exits `0` when healthy and `1` when unhealthy, which makes it directly usable as a container or monitoring health check. The shipped Docker Compose file uses it that way.

```bash
casci --status
```

## Shell Integration

| Flag | Description |
|------|-------------|
| `--shell completions [SHELL]` | Print shell completions |
| `--shell init [SHELL]` | Print shell init for `eval` |
| `--shell help` | Show shell help |

!!! warning "`--shell` is not implemented"
    Every form of `--shell`, including `--shell help`, prints `shell integration is not implemented yet` to stderr and exits non-zero. Nothing is written to stdout, so `eval "$(casci --shell init)"` evaluates an empty string and silently does nothing. The subcommands above are the planned surface, not current behavior.

## Server Configuration

| Flag | Description | Default |
|------|-------------|---------|
| `--mode {production\|development\|debug}` | Application mode | `production` |
| `--config DIR` | Config directory | OS default |
| `--data DIR` | Data directory | OS default |
| `--cache DIR` | Cache directory | OS default |
| `--log DIR` | Log directory | OS default |
| `--backup DIR` | Backup directory | OS default |
| `--pid FILE` | PID file path | OS default |
| `--address ADDR` | Listen address | `[::]` |
| `--port PORT` | Listen port | Random `64xxx`, `80` in a container |
| `--baseurl PATH` | URL path prefix | `/` |
| `--daemon` | Run as a daemon, detaching from the terminal | off |
| `--debug` | Enable debug mode: verbose logging and debug endpoints | off |
| `--color {auto\|yes\|no}` | Color output | `auto` |
| `--lang CODE` | Language for output | auto, from `LANG` |

Every directory flag creates its directory if it does not exist, then verifies it is writable and fails fast if it is not. See [Configuration](configuration.md) for the platform-specific defaults.

`--debug` is distinct from `--mode debug`: the flag turns on verbose logging for the current run, while the mode is an explicit opt-in that also changes application behavior.

!!! note "Four flags are accepted but not yet acted on"
    `--pid`, `--daemon`, `--baseurl`, and `--lang` are parsed and validated — `--baseurl` is rejected unless it starts with `/` — but nothing consumes them yet. No PID file is written, the process does not detach, output is not localized, and there is no URL path prefix. Run under a service manager for supervision, and serve CASCI at the root of its own hostname.

## Service Management

| Flag | Description |
|------|-------------|
| `--service start` | Start the server |
| `--service stop` | Stop the service |
| `--service restart` | Restart the service |
| `--service reload` | Reload configuration |
| `--service --install` | Install as a system service |
| `--service --uninstall` | Remove the system service |
| `--service --disable` | Disable the service |
| `--service --help` | Show service help |

```bash
casci --service --help
```

!!! note
    Only `--service --help` and `--service start` are wired up today. `start` directs you to run the binary without `--service`; the remaining subcommands report that they are not implemented yet. Use the platform installers described in [Installation](installation.md) in the meantime.

## Maintenance

| Flag | Description |
|------|-------------|
| `--maintenance backup [FILE]` | Create a backup |
| `--maintenance restore [FILE]` | Restore from a backup |
| `--maintenance update` | Apply maintenance updates |
| `--maintenance mode [SETTING]` | Read or change maintenance mode |
| `--maintenance setup` | Run first-time setup |
| `--maintenance --help` | Show maintenance help |

!!! note
    Only `--maintenance --help` does anything. `backup`, `restore`, `update`, `setup`, and a `mode` change all report that they are not implemented yet. Argument validation still runs first, so `restore` without a file and `mode` with an unknown value fail on the argument rather than on the missing implementation.

## Updates

| Flag | Description |
|------|-------------|
| `--update check` | Check for a newer release without installing |
| `--update yes` | Download and apply an update |
| `--update branch {stable\|beta\|daily}` | Select the update channel |
| `--update --help` | Show update help |

!!! note
    Only `--update --help` does anything. `check`, `yes`, and channel selection all report that they are not implemented yet. Update through your package manager or by replacing the binary.

## Color and Emoji Output

CASCI follows the [NO_COLOR](https://no-color.org/) standard. Precedence, highest first:

1. `--color=yes` or `--color=no`
2. `NO_COLOR` set to any non-empty value, which disables both color and emoji
3. Auto-detection from TTY state and `TERM`

There is no config-file setting for color; it is decided per run.

`TERM=dumb` disables all ANSI escapes, not only color. Neither `NO_COLOR` nor `TERM=dumb` changes output structure, table layout, or bold and underline styling.

```bash
# Plain output for logs and pipes
NO_COLOR=1 casci --status

# Force color even when piping
casci --status --color=yes
```

## Examples

```bash
# Start with defaults; no flags required
casci

# Start on a fixed port with a dedicated data directory
casci --port 64580 --data /srv/casci/data

# Development mode with verbose logging
casci --mode development --debug

# Bind to a private interface behind a reverse proxy
casci --address 127.0.0.1 --port 64580
```
