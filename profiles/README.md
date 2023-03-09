# Профили pprof
`base.server.pprof` и `result.server.pprof` - профили памяти для конфигурации с in-memory хранилищем и включенной проверкой подписи.

### Воспроизведение
```bash
# Запустите агент:
./cmd/agent/agent -k "xxx" -p 2s -r 2s

# Запустите сервер:
./cmd/server/server -k "xxx" -p "0.0.0.0:9090"

# Снимите профиль:
sleep 30 && curl -s http://localhost:9090/debug/pprof/heap > heap.out
```

### Сравнение результатов
```bash
# Из задания курса:
go tool pprof -top -diff_base=profiles/base.server.pprof profiles/result.server.pprof

# Более детальное в интерактивном режиме:
go tool pprof -nodefraction=0 -base ./profiles/base.server.pprof ./cmd/server/server ./profiles/result.server.pprof

# Для получение еще более подробной информации добавьте флаги `-nodefraction 0 -nodecount 100000`.
```

### Дополнительные опции
```bash
# Убрать отрицательные значения:
(pprof) drop_negative=true

# Показать дополнительную информацию в команде top:
(pprof) granularity=lines
```
