# metrics-collector
Репозиторий для практического трека «Go в DevOps»: реализация сбора (агент) и хранения (сервер) метрик.

## Задания
1.  [Инкремент 1 (агент)](./docs/tasks/increment1.md)
2.  [Инкремент 2 (сервер)](./docs/tasks/increment2.md)
3.  [Инкремент 3 (web framework)](./docs/tasks/increment3.md)
4.  [Инкремент 4 (JSON API)](./docs/tasks/increment4.md)
5.  [Инкремент 5 (переменные окружения)](./docs/tasks/increment5.md)
6.  [Инкремент 6 (сохранение данных на диск)](./docs/tasks/increment6.md)
7.  [Инкремент 7 (флаги командной строки)](./docs/tasks/increment7.md)
8.  [Инкремент 8 (поддержка сжатия)](./docs/tasks/increment8.md)
9.  [Инкремент 9 (подписывание передаваемых данных)](./docs/tasks/increment9.md)
10. [Инкремент 10 (проверка соединения с базой)](./docs/tasks/increment10.md)
11. [Инкремент 11 (сохранение данных в базе)](./docs/tasks/increment11.md)
12. [Инкремент 12 (отправка метрик списком)](./docs/tasks/increment12.md)
13. [Инкремент 13 (ошибки и логирование)](./docs/tasks/increment13.md)

## Разработка и тестирование
Для получения полного списка доступных команд выполните:
```bash
make help
```

### golangci-lint
В проекте используется `golangci-lint` для локальной разработки. Для установки линтера воспользуйтесь [официальной инструкцией](https://golangci-lint.run/usage/install/).

### pre-commit
В проекте используется `pre-commit` для запуска линтеров перед коммитом. Для установки утилиты воспользуйтесь [официальной инструкцией](https://pre-commit.com/#install), затем выполните команду:
```bash
make install-tools
```

### migrate
Для работы с миграциями БД необходимо установить утилиту [golang-migrate](https://github.com/golang-migrate/migrate):
```bash
go install -tags "postgres" github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Команды
Для добавления новой миграции выполните:
```bash
migrate create -ext sql -dir ./migrations -seq имя_миграции
```

Для применения миграций выполните команду:
```bash
migrate -database ${DATABASE_DSN} -path ./migrations up
```

Для возврата базы данных в первоначальное состояние выполните команду:
```bash
migrate -database ${DATABASE_DSN} -path ./migrations down -all
```

## Запуск сервера
Для запуска сервера, отвечающего за агрегирование и хранение метрик, выполните команду:
```bash
./cmd/server/server
```

### Настройка и значения по умолчанию
#### Опции командной строки
Для вывода списка доступных опций и их значений по умолчанию выполните команду:
```bash
./cmd/server/server --help
```

#### Переменные окружения сервера
(!) Переменные окружения имеют приоритет перед опциями командной строки.

```bash
# Адрес и порт, по которым доступно API сервера:
export ADDRESS=0.0.0.0:8080

# Интервал времени в секундах, по истечении которого текущие показания
# сервера сбрасываются на диск (значение 0 — делает запись синхронной):
export STORE_INTERVAL=300s

# Имя файла, где хранятся значения метрик.
# Пустое значение — отключает функцию записи на диск:
export STORE_FILE="/tmp/devops-metrics-db.json"

# Загружать или нет сохраненные значения метрик из файла при старте сервера:
export RESTORE=true

# Секретный ключ для генерации подписи (по умолчанию не задан):
export KEY=

# Полный URL для установления соединения с базой (по умолчанию не задан).
# Сервер автоматически запустит все необходимые миграции после установления соединения с базой.
# (!) Поддерживается только Postgres.
export DATABASE_DSN=

# Включить вывод отладочной информации.
export DEBUG=false
```

## Запуск агента
Для запуска агента, отвечающего за сбор и отправку метрик, выполните команду:
```bash
./cmd/agent/agent
```

### Настройка и значения по умолчанию
#### Опции командной строки
Для вывода списка доступных опций и их значений по умолчанию выполните команду:
```bash
./cmd/agent/agent --help
```

#### Переменные окружения агента
(!) Переменные окружения имеют приоритет перед опциями командной строки.

```bash
# Адрес и порт сервера, агрегирующего метрики:
export ADDRESS=0.0.0.0:8080

# Интервал опроса метрик (в секундах):
export POLL_INTERVAL=2s

# Интервал отправки метрик (в секундах):
export REPORT_INTERVAL=10s

# Секретный ключ для генерации подписи (по умолчанию не задан):
export KEY=

# Включить вывод отладочной информации.
export DEBUG=false
```

## Обновление шаблона
Чтобы получать обновления автотестов и других частей шаблона, выполните следующую команду:
```
git remote add -m main template https://github.com/yandex-praktikum/go-musthave-devops-tpl.git
```

Для обновления кода автотестов выполните команду:
```
git fetch template && git checkout template/main .github
```

## Лицензия
Copyright (c) 2022-2023 Alexander Kurbatov

Лицензировано по [GPLv3](LICENSE).
