#!/usr/bin/env bash
# Builds a universal macOS release. Run this on a macOS host with Xcode and
# the Wails/macOS native build dependencies installed.

set -Eeuo pipefail

repository_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
debug=false
skip_frontend_install=false

usage() {
  echo "Usage: $0 [--debug] [--skip-frontend-install]"
}

for argument in "$@"; do
  case "$argument" in
    --debug) debug=true ;;
    --skip-frontend-install) skip_frontend_install=true ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $argument" >&2; usage >&2; exit 2 ;;
  esac
done

for command in go node npm wails; do
  command -v "$command" >/dev/null || {
    echo "Required command not found on PATH: $command" >&2
    exit 1
  }
done

cd "$repository_root"

if [[ "$skip_frontend_install" == false ]]; then
  (cd frontend && npm ci)
fi

version="$(git describe --tags --always 2>/dev/null || echo dev)"
commit="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
build_date="$(date -u +%F)"
ldflags="-X main.Version=${version} -X main.Commit=${commit} -X main.BuildDate=${build_date}"

arguments=(-platform darwin/universal -ldflags "$ldflags")
if [[ "$debug" == true ]]; then
  arguments+=(-debug)
fi

wails build "${arguments[@]}"
