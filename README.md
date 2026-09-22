# netwatch

[![CI](https://github.com/Blizoman/netwatch/actions/workflows/ci.yml/badge.svg)](https://github.com/Blizoman/netwatch/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/Blizoman/netwatch.svg)](https://pkg.go.dev/github.com/Blizoman/netwatch)
[![Go Report Card](https://goreportcard.com/badge/github.com/Blizoman/netwatch)](https://goreportcard.com/report/github.com/Blizoman/netwatch)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

`netwatch` is a cross-platform CLI for network diagnostics. It monitors hosts,
TCP ports, DNS resolution, latency, packet loss and network paths
(traceroute), in one-shot or continuous modes, with human-readable or JSON
output.

```
── 2024-06-01T12:00:00Z ──
OK    dns         example.com      3ms   93.184.216.34
OK    tcp         example.com:443  18ms  connected to 93.184.216.34:443
OK    ping        example.com      21ms  loss=0.0% min=19ms avg=21ms max=24ms (unprivileged)
FAIL  traceroute  8.8.8.8          1.2s  destination not reached within max hops
1 hop listing indented under the summary...
4 ok, 1 failed (5 total)
```

## Features

- **DNS resolution** — measures lookup time and reports resolved addresses.
- **TCP connect checks** — measures handshake time to a `host:port`.
- **ICMP ping** — round-trip latency and packet loss, with automatic fallback
  from unprivileged to privileged sockets.
- **Traceroute** — hop-by-hop path discovery over ICMP.
- **One-shot (`check`) or continuous (`watch`) modes**, with graceful
  `Ctrl+C` cancellation.
- **Human-readable (colored) or JSON output** — JSON is newline-delimited,
  ready to pipe into `jq` or a log pipeline.
- **YAML config files** for a fixed set of targets, or ad-hoc `--target` flags
  for one-off checks.
- **Concurrent checks**, bounded by `--concurrency`.

## Install

### Go install

```sh
go install github.com/Blizoman/netwatch/cmd/netwatch@latest
```

### Download a release binary

Grab a prebuilt archive for your platform from the
[releases page](https://github.com/Blizoman/netwatch/releases).

### Docker

```sh
docker build -t netwatch .
docker run --rm --cap-add=NET_RAW netwatch check -t 1.1.1.1
```

`--cap-add=NET_RAW` is only needed for `ping`/`traceroute` checks (see
[Privileges](#privileges) below).

### Build from source

```sh
git clone https://github.com/Blizoman/netwatch.git
cd netwatch
make build   # binary at ./bin/netwatch
```

## Usage

### One-shot checks

```sh
# A single target, default checks (dns, tcp, ping — tcp only if a port is given)
netwatch check -t example.com:443

# Multiple targets, explicit checks per target
netwatch check -t example.com:443/dns,tcp,ping -t 1.1.1.1/ping,traceroute

# JSON output, piped into jq
netwatch check -t example.com:443 -o json | jq '.results[] | select(.success == false)'
```

The process exits with status `0` if every check succeeded, or `1` if any
check failed (useful in scripts and CI health checks) — the exit code is set
without printing a redundant error line, since the report already shows what
failed.

### Continuous monitoring

```sh
netwatch watch -t example.com:443 --interval 15s
```

Runs every check immediately, then again every `--interval`, until
interrupted with `Ctrl+C` (SIGINT) or SIGTERM, at which point it stops
cleanly instead of leaving in-flight checks running.

### Config files

For a fixed set of targets, use a YAML config instead of repeating `--target`:

```sh
netwatch watch --config configs/netwatch.example.yaml
```

See [`configs/netwatch.example.yaml`](configs/netwatch.example.yaml) for the
full schema:

```yaml
interval: 30s
timeout: 5s
ping_count: 5
max_hops: 30
concurrency: 8
output: human
no_color: false

targets:
  - host: example.com
    port: 443
    checks: [dns, tcp, ping]
  - host: 1.1.1.1
    checks: [ping, traceroute]
```

CLI flags (`--timeout`, `--output`, `--no-color`, etc.) override the
corresponding config file values when explicitly set.

### Target spec syntax

`--target`/`-t` (repeatable) accepts:

| Spec                          | Meaning                                       |
| ------------------------------ | ---------------------------------------------- |
| `host`                        | DNS + ping checks (no port, so no tcp check)  |
| `host:port`                   | DNS + tcp + ping checks                       |
| `host/check1,check2`          | Explicit checks, no port                      |
| `host:port/check1,check2`     | Explicit checks, with a port                  |

Valid checks: `dns`, `tcp`, `ping`, `traceroute`. `tcp` requires a port.

### All flags

```
netwatch check --help
netwatch watch --help
```

Common flags shared by both commands:

| Flag              | Default        | Description                                    |
| ------------------ | ---------------- | ------------------------------------------------- |
| `--config, -c`    |                | YAML config file (overrides `--target`)        |
| `--target, -t`    |                | Target spec, repeatable                        |
| `--checks`        | `dns,tcp,ping` | Default checks for `--target` entries          |
| `--timeout`       | `5s`           | Per-check timeout                              |
| `--ping-count`    | `5`            | ICMP echo requests per ping check               |
| `--max-hops`      | `30`           | Maximum TTL probed by traceroute               |
| `--concurrency`   | `8`            | Maximum concurrent checks                      |
| `--output, -o`    | `human`        | `human` or `json`                              |
| `--no-color`      | `false`        | Disable colored output                         |
| `--interval, -i`  | `30s`          | `watch` only: how often to re-run all checks   |

## Privileges

ICMP (`ping` and `traceroute`) needs elevated access to send/receive raw
network packets — the same requirement as the standard `ping`/`traceroute`
tools on any OS:

- **`ping`** first tries an unprivileged "ping socket" (no special
  permissions needed on Linux when `net.ipv4.ping_group_range` allows it, and
  out of the box on macOS), and automatically falls back to a privileged raw
  socket, which needs root or `CAP_NET_RAW`.
- **`traceroute`** always needs a privileged raw socket, because it has to
  observe ICMP "time exceeded" replies from routers along the path, which
  the unprivileged ping socket cannot see.

To grant `CAP_NET_RAW` without running as root on Linux:

```sh
sudo setcap cap_net_raw+ep ./bin/netwatch
```

If neither is available, `ping`/`traceroute` checks fail with a clear,
actionable error rather than hanging.

Traceroute currently targets IPv4 only.

## Exit codes

| Code | Meaning                                  |
| ---- | ----------------------------------------- |
| `0`  | Every check succeeded                    |
| `1`  | At least one check failed, or a usage/config error occurred |

## Development

```sh
make build     # build ./bin/netwatch
make test      # go test -race ./...
make lint      # golangci-lint run
make cover     # test coverage report
make docker    # build the Docker image
```

Project layout:

```
cmd/netwatch/        entrypoint
internal/cli/        cobra commands (check, watch, version)
internal/config/     YAML config loading + defaults
internal/target/     target spec parsing/validation
internal/checker/    dns, tcp, ping, traceroute implementations
internal/monitor/    concurrent check orchestration, one-shot/continuous
internal/output/     human and JSON renderers
internal/version/    build-time version metadata
configs/             example config file
```

## Versioning & releases

netwatch follows [Semantic Versioning](https://semver.org/). Releases are cut
by pushing a `vX.Y.Z` tag, which triggers
[GoReleaser](https://goreleaser.com/) via GitHub Actions to build
cross-platform binaries and publish a GitHub Release. See
[CHANGELOG.md](CHANGELOG.md) for release notes.

## Contributing

Issues and pull requests are welcome. Please run `make test lint` before
submitting.

## License

[MIT](LICENSE)
