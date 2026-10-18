#!/bin/sh
# The private state and lock belong to the managed SSH account on this host.
set -eu
export LC_ALL=C
export PATH=/usr/sbin:/sbin:/usr/bin:/bin
umask 077

action=$1 owner=$2 server=$3 token=$4 timeout=$5
shift 5
valid_id() {
    [ "${#1}" -eq 36 ] || return 1
    case "$1" in *[!a-fA-F0-9-]*) return 1;; esac
}
valid_id "$owner" && valid_id "$server" && valid_id "$token" || exit 64
for tool in flock stat nohup date sha256sum awk sleep rm cat mv id iptables iptables-save iptables-restore; do
    command -v "$tool" >/dev/null 2>&1 || exit 127
done
uid=$(id -u)
if [ "$uid" -ne 0 ] && ! command -v sudo >/dev/null 2>&1; then exit 126; fi
base="/tmp/nextgen-fw-$uid"
if [ ! -e "$base" ]; then
    mkdir -m 700 "$base" 2>/dev/null || [ -d "$base" ] || exit 74
fi
[ ! -L "$base" ] && [ "$(stat -c '%u:%a' "$base")" = "$uid:700" ] || exit 126
fw() {
    if [ "$uid" -eq 0 ]; then "$@"; else sudo -n "$@"; fi
}
revision() {
    rules=$(fw iptables -w 5 -S) || return 126
    printf '%s\n' "$rules" | sha256sum | { read -r digest rest; printf '%s' "$digest"; }
}
exec 9>"$base/lock"
flock -w 5 9 || exit 73

read_pending() {
    [ -f "$base/active" ] || return 1
    IFS=' ' read -r pending pending_owner pending_server expires seconds < "$base/active"
    valid_id "$pending" && valid_id "$pending_owner" && valid_id "$pending_server" || exit 74
    case "$expires:$seconds" in *[!0-9:]*) exit 74;; esac
}
require_owner() {
    [ "$pending_owner" = "$owner" ] && [ "$pending_server" = "$server" ] || exit 73
}
finish_pending() {
    rm -f "$base/active" "$base/$pending.rules" "$base/$pending.watchdog" "$base/$pending.ready"
}
apply_rules() {
    if [ "${1:-}" != "_MOVE" ]; then fw iptables -w 5 "$@"; return; fi
    [ "$#" -eq 4 ] || return 1
    # iptables-save data is never evaluated by the shell. Preserve every table
    # and line except the requested ordering of rules in the filter chain.
    candidate="$base/$token.candidate"
    if awk -v chain="$2" -v from="$3" -v dest="$4" '
        { lines[NR]=$0 }
        $0=="*filter" { in_filter=1 }
        in_filter && index($0,"-A " chain " ")==1 { rules[++count]=$0; positions[NR]=1 }
        $0=="COMMIT" { in_filter=0 }
        END {
            if (from<1 || from>count || dest<1 || dest>count) exit 1
            source=1; target=0
            for (line=1; line<=NR; line++) {
                if (!positions[line]) { print lines[line]; continue }
                target++
                if (target==dest) print rules[from]
                else { if (source==from) source++; print rules[source++] }
            }
        }
    ' "$snapshot" > "$candidate" && fw iptables-restore -w 5 < "$candidate"; then
        rm -f "$candidate"
        return 0
    fi
    rm -f "$candidate"
    return 1
}

case "$action" in
    status)
        if read_pending; then
            printf 'PENDING %s %s %s %s %s\n' "$pending" "$pending_owner" "$pending_server" "$expires" "$seconds"
        else
            printf 'NONE\n'
        fi
        digest=$(revision) || exit 126
        printf 'REVISION %s\nRULES\n' "$digest"
        fw iptables -w 5 -L -n -v --line-numbers
        ;;
    apply)
        expected=$1
        shift
        case "$timeout" in ''|*[!0-9]*) exit 64;; esac
        [ "$timeout" -ge 10 ] && [ "$timeout" -le 300 ] || exit 64
        if read_pending; then exit 73; fi
        if [ -n "$expected" ]; then
            case "$expected" in *[!a-f0-9]*) exit 64;; esac
            [ "${#expected}" -eq 64 ] || exit 64
            digest=$(revision) || exit 126
            [ "$digest" = "$expected" ] || exit 76
        fi
        snapshot="$base/$token.rules"
        fw iptables-save > "$snapshot" || exit 126
        # Verify restore authorization and parseability before any mutation.
        fw iptables-restore --test -w 5 < "$snapshot" || exit 126
        expires=$(( $(date +%s) + timeout ))
        printf '%s %s %s %s %s\n' "$token" "$owner" "$server" "$expires" "$timeout" > "$base/active.tmp"
        # The watchdog never shares fd 9 with the parent. It remains independent
        # of the SSH channel and uses the same lock before checking its token.
        cat > "$base/$token.watchdog" <<'NEXTGEN_WATCHDOG'
#!/bin/sh
set -eu
export PATH=/usr/sbin:/sbin:/usr/bin:/bin
base=$1 token=$2 expiry=$3 uid=$4
fw() { if [ "$uid" -eq 0 ]; then "$@"; else sudo -n "$@"; fi; }
: > "$base/$token.ready"
remaining=$(( expiry - $(date +%s) ))
if [ "$remaining" -gt 0 ]; then sleep "$remaining"; fi
while :; do
    (
        exec 9>"$base/lock"
        flock 9
        [ -f "$base/active" ] || exit 0
        IFS=' ' read -r current owner server expires seconds < "$base/active"
        [ "$current" = "$token" ] || exit 0
        if fw iptables-restore -w 5 < "$base/$token.rules"; then
            rm -f "$base/active" "$base/$token.rules" "$base/$token.watchdog" "$base/$token.ready"
            exit 0
        fi
        exit 1
    ) && exit 0
    sleep 1
done
NEXTGEN_WATCHDOG
        nohup sh "$base/$token.watchdog" "$base" "$token" "$expires" "$uid" </dev/null >/dev/null 2>&1 9>&- &
        child=$!
        attempts=0
        while [ ! -f "$base/$token.ready" ]; do
            if ! kill -0 "$child" 2>/dev/null || [ "$attempts" -ge 40 ]; then
                # No mutation has run, so removing this marker is safe.
                pending=$token
                finish_pending
                exit 74
            fi
            attempts=$(( attempts + 1 ))
            sleep 0.05
        done
        # Never start a mutation after the protection window has elapsed.
        mv "$base/active.tmp" "$base/active"
        if [ "$(date +%s)" -ge "$expires" ]; then
            pending=$token
            finish_pending
            exit 75
        fi
        if [ "$#" -gt 0 ] && ! apply_rules "$@"; then
            if fw iptables-restore -w 5 < "$snapshot"; then
                pending=$token
                finish_pending
            fi
            exit 74
        fi
        printf 'BACKUP %s %s %s\n' "$token" "$expires" "$timeout"
        cat "$snapshot"
        ;;
    confirm)
        read_pending || exit 75
        require_owner
        [ "$pending" = "$token" ] && [ "$(date +%s)" -lt "$expires" ] || exit 75
        finish_pending
        printf 'OK\n'
        ;;
    rollback)
        if read_pending; then
            require_owner
            [ "$pending" = "$token" ] || exit 75
        fi
        restore="$base/$token.restore"
        cat > "$restore"
        if ! fw iptables-restore -w 5 < "$restore"; then
            rm -f "$restore"
            exit 74
        fi
        rm -f "$restore"
        if read_pending; then finish_pending; fi
        printf 'OK\n'
        ;;
    *) exit 64;;
esac
