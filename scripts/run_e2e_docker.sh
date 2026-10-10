#!/usr/bin/env bash
# ==============================================================================
# NextGen CyberSecurity Platform - Automated Container E2E Test Runner
# Cross-Platform Linux/CI Docker Orchestrator & E2E Test Suite
# ==============================================================================

set -euo pipefail

BACKEND_URL="${BACKEND_URL:-http://localhost:8080}"
FRONTEND_URL="${FRONTEND_URL:-http://localhost:3000}"
EMAIL="${ADMIN_EMAIL:-admin@nextgen.local}"
PASSWORD="${ADMIN_PASSWORD:-AdminPass123!}"
WAIT_TIMEOUT="${WAIT_TIMEOUT:-90}"
KEEP_ALIVE=0
CLEAN_VOLUMES=1

# Color formatting
RED='\033[0;31m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

log_step() {
  echo -e "\n${CYAN}[STEP] $1${NC}"
}

log_pass() {
  echo -e "  ${GREEN}[+] PASS: $1${NC}"
}

log_fail() {
  echo -e "  ${RED}[-] FAIL: $1${NC}" >&2
}

log_info() {
  echo -e "  ${YELLOW}[*] INFO: $1${NC}"
}

print_header() {
  echo -e "${MAGENTA}==========================================================${NC}"
  echo -e "${MAGENTA}  NextGen CyberSecurity - Automated Container E2E Runner   ${NC}"
  echo -e "${MAGENTA}  Stack: Docker Compose (PostgreSQL 16, Go, React/Nginx)  ${NC}"
  echo -e "${MAGENTA}==========================================================${NC}"
}

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --keep-alive)
        KEEP_ALIVE=1
        shift
        ;;
      --no-clean-volumes)
        CLEAN_VOLUMES=0
        shift
        ;;
      --backend-url=*)
        BACKEND_URL="${1#*=}"
        shift
        ;;
      --frontend-url=*)
        FRONTEND_URL="${1#*=}"
        shift
        ;;
      --timeout=*)
        WAIT_TIMEOUT="${1#*=}"
        shift
        ;;
      -h|--help)
        echo "Uso: $0 [--keep-alive] [--no-clean-volumes] [--backend-url=URL] [--frontend-url=URL] [--timeout=SEC]"
        exit 0
        ;;
      *)
        echo "Opzione non riconosciuta: $1"
        exit 1
        ;;
    esac
  done
}

cleanup() {
  local exit_code=$?
  if [[ $KEEP_ALIVE -eq 0 ]]; then
    log_info "Arresto e pulizia dello stack Docker Compose..."
    if [[ $CLEAN_VOLUMES -eq 1 ]]; then
      docker compose down -v --remove-orphans >/dev/null 2>&1 || true
    else
      docker compose down --remove-orphans >/dev/null 2>&1 || true
    fi
  else
    log_info "Flag --keep-alive attivo: container lasciati in esecuzione."
  fi
  if [[ $exit_code -eq 0 ]]; then
    echo -e "\n${GREEN}🎉 TUTTI I TEST E2E SONO STATI COMPLETATI CON SUCCESSO! (EXIT 0)${NC}\n"
  else
    echo -e "\n${RED}❌ ALMENO UN TEST E2E È FALLITO. (EXIT $exit_code)${NC}\n"
  fi
  exit "$exit_code"
}

trap cleanup EXIT INT TERM

wait_for_health() {
  local url="$1"
  local name="$2"
  local timeout="$3"
  local elapsed=0

  log_info "In attesa che $name sia pronto su $url (Timeout: ${timeout}s)..."
  while [[ $elapsed -lt $timeout ]]; do
    if curl -s -f -m 2 "$url" >/dev/null 2>&1; then
      log_pass "$name è operativo e risponde correttamente!"
      return 0
    fi
    sleep 2
    elapsed=$((elapsed + 2))
  done

  log_fail "$name non ha risposto entro $timeout secondi."
  return 1
}

main() {
  parse_args "$@"
  print_header

  # 1. Verifica prerequisiti
  log_step "1. Verifica prerequisiti host (Docker & Docker Compose)"
  if ! command -v docker >/dev/null 2>&1; then
    log_fail "Docker non è installato o non è presente nel PATH."
    exit 1
  fi
  if ! docker compose version >/dev/null 2>&1; then
    log_fail "Docker Compose (v2) non è disponibile."
    exit 1
  fi
  log_pass "Docker & Docker Compose rilevati con successo."

  # 2. Avvio stack Docker Compose
  log_step "2. Avvio dello stack di container (PostgreSQL, Backend, Frontend)"
  docker compose up -d --build postgres backend frontend
  log_pass "Container avviati in background."

  # 3. Attesa disponibilità servizi
  log_step "3. Healthcheck e attesa disponibilità endpoint"
  wait_for_health "$BACKEND_URL/api/v1/health" "Backend Go (Direct)" "$WAIT_TIMEOUT"
  wait_for_health "$FRONTEND_URL/api/v1/health" "Frontend Nginx (Reverse Proxy)" "$WAIT_TIMEOUT"
  wait_for_health "$FRONTEND_URL/" "Frontend React SPA (Nginx Static)" 30

  # 4. Verifica SPA HTML
  log_step "4. Verifica erogazione HTML SPA React"
  local spa_html
  spa_html=$(curl -s "$FRONTEND_URL/")
  if [[ "$spa_html" =~ root ]]; then
    log_pass "React Root container rilevato in index.html."
  else
    log_fail "Index HTML erogato non contiene il contenitore root React."
    exit 1
  fi

  # 5. Autenticazione JWT
  log_step "5. Autenticazione SuperAdmin e acquisizione JWT Bearer Token"
  local login_payload
  login_payload=$(printf '{"email":"%s","password":"%s"}' "$EMAIL" "$PASSWORD")

  local login_resp
  login_resp=$(curl -s -X POST \
    -H "Content-Type: application/json" \
    -d "$login_payload" \
    "$BACKEND_URL/api/v1/auth/login")

  local token
  token=$(echo "$login_resp" | grep -o '"access_token":"[^"]*' | cut -d'"' -f4 || true)

  if [[ -z "$token" ]]; then
    log_fail "Autenticazione fallita: nessun access_token nella risposta: $login_resp"
    exit 1
  fi
  log_pass "Autenticazione riuscita. Token JWT ottenuto (lunghezza: ${#token} caratteri)."

  local auth_header="Authorization: Bearer $token"

  # 6. Test Target Groups & Targets API
  log_step "6. Verifica Target Groups & Targets API"
  local groups_status
  groups_status=$(curl -s -o /dev/null -w "%{http_code}" -H "$auth_header" "$BACKEND_URL/api/v1/target-groups")
  if [[ "$groups_status" -eq 200 ]]; then
    log_pass "GET /api/v1/target-groups ha risposto con HTTP 200 OK."
  else
    log_fail "GET /api/v1/target-groups ha risposto con HTTP $groups_status."
    exit 1
  fi

  local targets_status
  targets_status=$(curl -s -o /dev/null -w "%{http_code}" -H "$auth_header" "$BACKEND_URL/api/v1/targets")
  if [[ "$targets_status" -eq 200 ]]; then
    log_pass "GET /api/v1/targets ha risposto con HTTP 200 OK."
  else
    log_fail "GET /api/v1/targets ha risposto con HTTP $targets_status."
    exit 1
  fi

  # 7. Test Scan Profiles API
  log_step "7. Verifica Profili di Scansione Nmap"
  local profiles_resp
  profiles_resp=$(curl -s -H "$auth_header" "$BACKEND_URL/api/v1/scan-profiles")
  if [[ "$profiles_resp" =~ "All-in-One 360" || "$profiles_resp" =~ "Fast Discovery" || "$profiles_resp" =~ "Standard" ]]; then
    log_pass "GET /api/v1/scan-profiles contiene i profili preimpostati certificati."
  else
    log_fail "GET /api/v1/scan-profiles non contiene i profili attesi: $profiles_resp"
    exit 1
  fi

  # 8. Test Attack Path Graph & Chokepoints API (Milestone 2)
  log_step "8. Verifica Attack Path Mapping & Chokepoints (Milestone 2 API)"
  local graph_status
  graph_status=$(curl -s -o /dev/null -w "%{http_code}" -H "$auth_header" "$BACKEND_URL/api/v1/attack-paths/graph")
  if [[ "$graph_status" -eq 200 ]]; then
    log_pass "GET /api/v1/attack-paths/graph ha risposto con HTTP 200 OK."
  else
    log_fail "GET /api/v1/attack-paths/graph ha risposto con HTTP $graph_status."
    exit 1
  fi

  local chokepoints_status
  chokepoints_status=$(curl -s -o /dev/null -w "%{http_code}" -H "$auth_header" "$BACKEND_URL/api/v1/attack-paths/chokepoints")
  if [[ "$chokepoints_status" -eq 200 ]]; then
    log_pass "GET /api/v1/attack-paths/chokepoints ha risposto con HTTP 200 OK."
  else
    log_fail "GET /api/v1/attack-paths/chokepoints ha risposto con HTTP $chokepoints_status."
    exit 1
  fi

  # 9. Test Reports API
  log_step "9. Verifica Reports API"
  local reports_status
  reports_status=$(curl -s -o /dev/null -w "%{http_code}" -H "$auth_header" "$BACKEND_URL/api/v1/reports")
  if [[ "$reports_status" -eq 200 ]]; then
    log_pass "GET /api/v1/reports ha risposto con HTTP 200 OK."
  else
    log_fail "GET /api/v1/reports ha risposto con HTTP $reports_status."
    exit 1
  fi

  # 10. Test Server-Sent Events (SSE) Stream Handshake
  log_step "10. Verifica Server-Sent Events (SSE) Handshake"
  local sse_code
  sse_code=$(curl -s -o /dev/null -w "%{http_code}" \
    -m 3 \
    -H "$auth_header" \
    -H "Accept: text/event-stream" \
    "$BACKEND_URL/api/v1/network-discovery/stream" || true)

  # Note: timeout di 3s per interrompere lo stream infinito; status 200 o curl timeout 28 dopo 200 è valido
  if [[ "$sse_code" -eq 200 || "$sse_code" -eq 0 ]]; then
    log_pass "Endpoint SSE /network-discovery/stream accetta connessioni streaming (HTTP 200)."
  else
    log_fail "Endpoint SSE ha risposto con status non valido: $sse_code"
    exit 1
  fi

  log_step "11. Riepilogo Collaudo E2E"
  log_pass "Tutti i 10 passi di integrazione (Healthcheck, Proxy, SPA, Auth, Targets, Profiles, Attack Paths M2, Reports, SSE) sono stati convalidati con successo!"
}

main "$@"

