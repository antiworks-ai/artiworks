# CLI Config Loader MVP Design

## Goal

Provide the first real `artiworks` CLI bootstrap and a tested TOML config loader so later API server, control plane, and TUI entrypoints can share the same process wiring.

## Scope

- Add a lightweight TOML config loader under `internal/app/configloader`.
- Add a stdlib-only CLI command runner under `internal/app/cli`.
- Keep `cmd/artiworks/main.go` as a thin process entrypoint.
- Support `version` and `status` commands now; reserve `serve`, `run`, and TUI commands for later design points.

## Config Resolution

Config resolution must be deterministic:

1. Explicit CLI path via `--config` loads that one file.
2. `ARTIWORKS_CONFIG` loads that one file when no explicit path is supplied.
3. Default layered config loads `$ARTIWORKS_HOME/config.toml` first, then `.artiworks/config.toml` from the working directory as an override.

Missing default files are allowed and produce a zero-value `config.AppConfig`. Missing explicit or env-selected config files are errors.

Project config overrides user config by deep-merging TOML maps before decoding into `config.AppConfig`, preserving explicit false and zero values in override files.

## CLI Behavior

- `artiworks version` prints build metadata to stdout.
- `artiworks status --config <path> --output json` loads config, builds the app wiring, and prints machine-readable status.
- `status` must not contact providers. It validates that the config can compose
  local runtime wiring, including local secret resolution for configured
  provider credentials.
- Diagnostics and errors go to stderr. Program output goes to stdout.
- `main` owns `os.Exit`; command code returns exit codes for testability.

## Dependencies

Use `github.com/pelletier/go-toml/v2` only for TOML parsing. Do not introduce Cobra/Viper in this design point.
