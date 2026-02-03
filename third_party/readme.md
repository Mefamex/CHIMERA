# THIRD PARTY ASSETS — MANUAL CHECKLIST

This repository keeps a small set of hand-managed third-party assets under `third_party/`.
This README is the canonical, human-readable guide describing what belongs here and the
rules to follow when adding/updating files.


## PURPOSE
- Centralize small, app-specific external assets (CSS, tiny JS helpers, drivers) so the
  project can reference them from a consistent location.

## RULES
- Prefer package managers or CDNs for large or frequently-updated libraries.
- Commit only small assets with clear source and permissive license.
- Do not add large binaries to git history; use `scripts/download_third_party.sh`, Git LFS,
  or an artifact repository instead.

## THIRD-PARTY FOLDERS

- `web/static/third_party/` — Third-party static assets for the web interface.
