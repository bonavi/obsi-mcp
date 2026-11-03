# obsi-mcp

MCP-сервер над git-клоном базы Obsidian. Структура и стандарты кода как у серверов Tasks и SimpleVPN: `server` + подмодуль `pkg`, см. `CLAUDE.md`. Архитектура и решения: заметка «Claude-агент для Obsidian» в базе.

Инструменты: `list`, `read`, `search`, `write`, `replace`, `append`, `move`, `commit`. Удаления нет намеренно.

После клонирования: `git submodule update --init`.

```bash
cd server && go test -race ./...
./scripts/e2e.sh
```

`scripts/e2e.sh` поднимает контейнер на временном репозитории. Настоящую базу в проверки не подключать.

Переменные окружения обязательные, список в `compose.test.yaml`.

Вход пока по статическому токену `AUTH_TOKEN`: OAuth ещё не сделан, наружу сервер не публиковать.
