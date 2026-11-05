#!/usr/bin/env bash
# Сквозная проверка контейнера на временном репозитории.
set -euo pipefail
cd "$(dirname "$0")/.."
export E2E_DIR="$(mktemp -d)" E2E_SIGNING_KEY="$(openssl rand -hex 32)"
E2E_PASSWORD="$(openssl rand -hex 16)"
export E2E_PASSWORD_HASH="$(cd server && printf '%s\n' "$E2E_PASSWORD" | go run ./cmd/hashPassword)"
URL=http://127.0.0.1:18080/
fail=0
check() { if [ "$2" = "$3" ]; then echo "ok   $1"; else echo "FAIL $1: got [$2] want [$3]"; fail=1; fi; }
compose() { docker compose -f compose.test.yaml "$@"; }
trap 'compose logs --no-color | tail -20 > "$E2E_DIR/logs.txt"; compose down -v >/dev/null 2>&1' EXIT

G="git -c user.name=mac -c user.email=mac@example.invalid"
git init -q --bare -b main "$E2E_DIR/bare.git"
$G clone -q "$E2E_DIR/bare.git" "$E2E_DIR/mac" 2>/dev/null
cd "$E2E_DIR/mac"; git checkout -q -b main
mkdir Здоровье; echo тайна > Здоровье/x.md; printf 'старт\nлог\n' > note.md; echo конфликт > c.md
$G add -A; $G commit -qm init; git push -q origin main
cd - >/dev/null
git clone -q -b main "$E2E_DIR/bare.git" "$E2E_DIR/server"
git -C "$E2E_DIR/server" remote set-url origin /remote/bare.git
mkdir "$E2E_DIR/state"; chmod -R a+rwX "$E2E_DIR"

compose up -d --build --quiet-pull 2>&1 | tail -1
for _ in $(seq 40); do curl -sf -o /dev/null http://127.0.0.1:18090/livez && break; sleep 0.5; done

code() { curl -s -o /dev/null -w '%{http_code}' -X POST "$URL" -H "Authorization: Bearer $1" -H 'Content-Type: application/json' -H 'Accept: application/json, text/event-stream' -d '{"jsonrpc":"2.0","id":1,"method":"ping"}'; }
check "неверный токен" "$(code wrong)" 401
check "401 указывает на метаданные" "$(curl -si -X POST "$URL" -d '{}' | grep -ci 'resource_metadata=')" 1
check "метаданные ресурса" "$(curl -s http://127.0.0.1:18080/.well-known/oauth-protected-resource | python3 -c 'import json,sys;print(json.load(sys.stdin)["authorization_servers"][0])')" http://127.0.0.1:18080

# Настоящий вход: регистрация, форма, код, токены
TOKENS=$(E2E_PASSWORD="$E2E_PASSWORD" python3 scripts/oauthLogin.py http://127.0.0.1:18080)
E2E_TOKEN=$(printf '%s' "$TOKENS" | python3 -c 'import json,sys;print(json.load(sys.stdin)["access_token"])')
REFRESH=$(printf '%s' "$TOKENS" | python3 -c 'import json,sys;print(json.load(sys.stdin)["refresh_token"])')
check "токен от входа принят" "$(code "$E2E_TOKEN")" 200

req() { curl -s "$URL" -H "Authorization: Bearer $E2E_TOKEN" -H 'Content-Type: application/json' -H 'Accept: application/json, text/event-stream' -d "$1"; }
tool() { req "$(python3 -c 'import json,sys;print(json.dumps({"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":sys.argv[1],"arguments":json.loads(sys.argv[2])}}))' "$1" "$2")"; }
field() { python3 -c 'import json,sys;r=json.load(sys.stdin)["result"];print(r.get("structuredContent",{}).get(sys.argv[1],""))' "$1"; }
iserr() { python3 -c 'import json,sys;print(json.load(sys.stdin)["result"].get("isError",False))'; }

H=$(tool read '{"path":"note.md"}' | field hash)
EDIT=$(printf '{"path":"note.md","old_text":"старт","new_text":"правка","expected_hash":"%s"}' "$H")
STALE=$(printf '{"path":"note.md","old_text":"лог","new_text":"x","expected_hash":"%s"}' "$H")
check "правка" "$(tool replace "$EDIT" | iserr)" False
check "правка на диске" "$(head -1 "$E2E_DIR/server/note.md")" правка
check "устаревший хэш" "$(tool replace "$STALE" | iserr)" True
check "после устаревшего хэша заметка цела" "$(sed -n 2p "$E2E_DIR/server/note.md")" лог
check "чтение запрещённой папки" "$(tool read '{"path":"Здоровье/x.md"}' | iserr)" True
check "выход за базу" "$(tool read '{"path":"../../etc/passwd.md"}' | iserr)" True
check "запись в .git" "$(tool write '{"path":".git/hooks/x.md","content":"x"}' | iserr)" True
check "поиск не видит тайну" "$(tool search '{"query":"тайна"}' | field hits | tr -d "[]")" ""
check "коммит от Claude" "$(tool commit '{"message":"правка заметки"}' | field committed)" True
check "сервер запушил" "$(git -C "$E2E_DIR/bare.git" log -1 --format=%s main)" "правка заметки"

tool append '{"path":"note.md","text":"забыт коммит"}' >/dev/null
sleep 10
check "страховочный автокоммит" "$(git -C "$E2E_DIR/bare.git" log -1 --format=%s main | cut -c1-5)" "auto:"

# Конфликт: Mac и сервер правят одну строку между синками.
compose pause >/dev/null 2>&1
cd "$E2E_DIR/mac"; git pull -q origin main; echo от-мака > c.md; $G commit -qam mac; git push -q origin main; MAC=$(git rev-parse HEAD); cd - >/dev/null
echo от-сервера > "$E2E_DIR/server/c.md"
git -C "$E2E_DIR/server" add -A
git -C "$E2E_DIR/server" -c user.name=s -c user.email=s@s.invalid commit -qm "правка сервера"
chmod -R a+rwX "$E2E_DIR/server/.git"
compose unpause >/dev/null 2>&1
sleep 5
check "main не тронут конфликтом" "$(git -C "$E2E_DIR/bare.git" rev-parse main)" "$MAC"
B=$(git -C "$E2E_DIR/bare.git" for-each-ref --format='%(refname:short)' 'refs/heads/claude/conflict-*' | head -1)
check "правка в ветке конфликта" "$(git -C "$E2E_DIR/bare.git" show "$B:c.md" 2>/dev/null)" от-сервера
check "о конфликте сообщили" "$(tool append '{"path":"note.md","text":"x"}' | field notice | grep -c "$B")" 1

# Защита контейнера.
compose exec -T obsi-mcp sh -c 'id -u' | { read u; check "не root" "$u" 1000; }
check "корень только для чтения" "$(compose exec -T obsi-mcp sh -c 'touch /x 2>/dev/null && echo w || echo ro')" ro

# Перезапуск: вход переживает его, файл сессий закрыт и не попал в клон
compose restart >/dev/null 2>&1
for _ in $(seq 40); do curl -sf -o /dev/null http://127.0.0.1:18090/livez && break; sleep 0.5; done
check "access-токен после перезапуска" "$(code "$E2E_TOKEN")" 200
NEW=$(curl -s http://127.0.0.1:18080/oauth/token -d grant_type=refresh_token --data-urlencode "refresh_token=$REFRESH" | python3 -c 'import json,sys;print(json.load(sys.stdin).get("refresh_token",""))')
check "refresh после перезапуска" "$([ -n "$NEW" ] && echo yes)" yes
check "повтор refresh отклонён" "$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:18080/oauth/token -d grant_type=refresh_token --data-urlencode "refresh_token=$REFRESH")" 400
check "после повтора вход отозван" "$(code "$E2E_TOKEN")" 401
check "токены не в файле" "$(grep -c "$NEW" "$E2E_DIR/state/oauth.json" || true)" 0
check "файла сессий нет в клоне" "$(git -C "$E2E_DIR/server" ls-files | grep -c oauth || true)" 0

LEAK=$(compose logs --no-color | grep -c -E "правка|тайна|забыт|$E2E_PASSWORD|$REFRESH" || true)
check "журнал без содержимого" "$LEAK" 0
exit $fail
