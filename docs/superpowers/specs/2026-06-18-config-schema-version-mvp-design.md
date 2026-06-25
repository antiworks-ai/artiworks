# Config Schema Version MVP Design

## Goal

Add the v1 config schema version marker described by the architecture docs.

## Scope

Included:

- `config.AppConfig` exposes top-level `version`.
- The generated root `schema.json` includes `version`.
- Config files that omit `version` default to the current config schema version after loading.
- Missing config files still return the zero-value config.

Excluded:

- Rejecting unsupported future versions.
- Config migrations.
- Splitting `schema.json` into config/API/event schema files.

## Design Notes

The version field is a compatibility marker, not a migration system yet. The loader should preserve explicit versions and only default loaded config documents that omit the field.
