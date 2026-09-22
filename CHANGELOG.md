# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added

- Initial release: DNS, TCP, ICMP ping and traceroute checks.
- One-shot (`check`) and continuous (`watch`) monitoring modes.
- Human-readable colored output and newline-delimited JSON output.
- YAML configuration files, with CLI flags able to override individual settings.
- Concurrent, bounded check execution with graceful `Ctrl+C` cancellation.
- Docker image and multi-platform (Linux/macOS/Windows, amd64/arm64) release builds.
