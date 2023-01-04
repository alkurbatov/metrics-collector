# Инкремент 12
## Сервер
Добавьте новый хендлер `POST /updates/`, принимающий в теле запроса множество метрик в формате `[]Metrics` (списка метрик).

## Агент
Научите агент работать с использованием нового `API` (отправлять метрики батчами).

Стоит помнить, что:
- нужно соблюдать обратную совместимость;
- отправлять пустые батчи не нужно;
- вы умеете сжимать контент по алгоритму `gzip`;
- изменение в базе можно выполнять в рамках одной транзакции/одного запроса.