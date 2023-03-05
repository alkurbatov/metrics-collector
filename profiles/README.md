# Профили pprof

## `base.server.pprof`
Профиль памяти для конфигурации с in-memory хранилищем и включенной проверкой подписи.

Для воспроизведения:
```bash
# Запустите агент:
./cmd/agent/agent -k "xxx" -p 2s -r 2s

# Запустите сервер:
./cmd/server/server -k "xxx" -p "0.0.0.0:9090"

# Снимите профиль:
sleep 30 && curl -s http://localhost:9090/debug/pprof/heap > heap.out
```
