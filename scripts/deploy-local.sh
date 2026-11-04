#!/usr/bin/env bash
# Build the plugin and load it into the local Grafana (aiden-grafana, which mounts this repo's
# dist/; see ../aiden-assistant/scripts/container-up.sh).
#
#   scripts/deploy-local.sh          # rebuild and restart the plugin's backend
#   scripts/deploy-local.sh --full   # also restart Grafana, so browsers load the new frontend
#
# Grafana reads the plugin's version when it starts and serves module.js?_cache=<version> with
# an hour of browser caching. Each build here gets its own version, but only a Grafana restart
# picks it up: after the default (fast) deploy, hard-reload the browser to see frontend changes.
#
# Never restart aiden-grafana on its own (container stop/start): it comes back on a new IP, and
# the MCP server, forecasting service and Prometheus keep the old one. --full restarts it with
# container-up.sh, which reconnects them.
#
# Options (environment variables):
#   GRAFANA_NAME   Grafana's container. Default: aiden-grafana.
#   GRAFANA_PORT   Grafana's port on localhost. Default: 3333.
#   AIDEN_DIR      The aiden-assistant checkout, for --full. Default: ../aiden-assistant.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
GRAFANA_NAME="${GRAFANA_NAME:-aiden-grafana}"
GRAFANA_PORT="${GRAFANA_PORT:-3333}"
AIDEN_DIR="${AIDEN_DIR:-$ROOT/../aiden-assistant}"
PLUGIN_ID=grafana-object-datasource

log() { printf '\033[1;35m▸\033[0m %s\n' "$*"; }
die() { printf '\033[1;31m✗\033[0m %s\n' "$*" >&2; exit 1; }

FULL=0
case "${1:-}" in
  --full) FULL=1 ;;
  "") ;;
  *) die "Unknown option: $1 (use --full or nothing)" ;;
esac

cd "$ROOT"
if command -v yarn >/dev/null; then YARN=(yarn); else YARN=(npx --yes yarn@1.22.22); fi
[[ -d node_modules ]] || { log "Installing dependencies"; "${YARN[@]}" install --frozen-lockfile; }

log "Building frontend"
"${YARN[@]}" -s build >/dev/null

# A version of its own for each build, so a restarted Grafana serves a module.js browsers haven't cached
python3 - "$ROOT/dist/plugin.json" <<'PY'
import json, sys, time
path = sys.argv[1]
meta = json.load(open(path))
base = meta["info"]["version"].split("-dev.")[0]
meta["info"]["version"] = f"{base}-dev.{time.strftime('%Y%m%d%H%M%S', time.gmtime())}"
json.dump(meta, open(path, "w"), indent=2)
PY

log "Building backend for linux/arm64"
MAGE="$(command -v mage || echo "$(go env GOPATH)/bin/mage")"
[[ -x "$MAGE" ]] || die "mage is not installed: go install github.com/magefile/mage@latest"
"$MAGE" build:linuxARM64 >/dev/null

if [[ "$FULL" == "1" ]]; then
  [[ -x "$AIDEN_DIR/scripts/container-up.sh" ]] || die "No container-up.sh in $AIDEN_DIR (set AIDEN_DIR)"
  log "Restarting Grafana and the services around it"
  SKIP_BUILD=1 "$AIDEN_DIR/scripts/container-up.sh"
  exit 0
fi

container inspect "$GRAFANA_NAME" 2>/dev/null | grep -q '"state" : "running"' ||
  die "$GRAFANA_NAME isn't running. Start it with $AIDEN_DIR/scripts/container-up.sh"

# Grafana starts the plugin's backend again when its process exits
log "Restarting the plugin's backend"
container exec "$GRAFANA_NAME" pkill -f "gpx_object-datasource" || true
for _ in $(seq 1 20); do
  sleep 0.5
  container exec "$GRAFANA_NAME" pgrep -f "gpx_object-datasource" >/dev/null && break
done

STATUS="$(curl -s "http://127.0.0.1:$GRAFANA_PORT/api/plugins/$PLUGIN_ID/settings" |
  python3 -c 'import json,sys; print(json.load(sys.stdin)["info"]["version"])' 2>/dev/null || echo "unknown")"
log "Deployed. Grafana is serving frontend version $STATUS; hard-reload the browser (Cmd+Shift+R) for frontend changes, or use --full"
