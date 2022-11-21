# go-musthave-devops-tpl

Репозитория для практического трека «Go в DevOps».

## Задания
1. [Инкремент 1 (агент)](./docs/tasks/increment1.md)
1. [Инкремент 2 (сервер)](./docs/tasks/increment2.md)

## Обновление шаблона

Чтобы получать обновления автотестов и других частей шаблона, выполните следующую команду:
```
git remote add -m main template https://github.com/yandex-praktikum/go-musthave-devops-tpl.git
```

Для обновления кода автотестов выполните команду:
```
git fetch template && git checkout template/main .github
```
