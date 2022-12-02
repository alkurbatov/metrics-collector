# metrics-collector
Репозиторий для практического трека «Go в DevOps»: реализация сбора (агент) и хранения (сервер) метрик.

## Задания
1. [Инкремент 1 (агент)](./docs/tasks/increment1.md)
2. [Инкремент 2 (сервер)](./docs/tasks/increment2.md)
3. [Инкремент 3 (web framework)](./docs/tasks/increment3.md)

## Разработка и тестирование
Для получения списка доступных команд выполните следующую команду:
```bash
make help
```

## Запуск
Для запуска сервера, отвечающего за агрегирование и хранение метрик, выполните команду:
```bash
./cmd/server/server
```

Для запуска агента, отвечающего за сбор и отправку метрик, выполните команду:
```bash
./cmd/agent/agent
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
