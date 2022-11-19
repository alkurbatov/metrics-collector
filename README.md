# go-musthave-devops-tpl

Репозитория для практического трека «Go в DevOps».

## Обновление шаблона

Чтобы получать обновления автотестов и других частей шаблона, выполните следующую команду:
```
git remote add -m main template https://github.com/yandex-praktikum/go-musthave-devops-tpl.git
```

Для обновления кода автотестов выполните команду:
```
git fetch template && git checkout template/main .github
```
