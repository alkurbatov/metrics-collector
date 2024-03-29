# Инкремент 4
1. Дополните API сервера, которое позволяет принимать метрики в формате JSON.  
   При реализации задействуйте одну из распространённых библиотек, например `encoding/json`.
   Обмен с сервером организуйте с использованием следующей структуры:
    ```golang
  type Metrics struct {
        ID    string   `json:"id"`              // имя метрики
        MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
        Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
        Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
  }
   ```
   Для передачи метрик на сервер используйте `Content-Type: application/json`. В теле запроса — описанный выше JSON. Передача через `POST update/`. В теле ответа отправляйте JSON той же структуры с актуальным (изменённым) значением `Value`.
   Для получения метрик с сервера используйте `Content-Type: application/json`. В теле запроса — описанный выше JSON (заполняйте только ID и MType). В ответ получайте такой же JSON, но с уже заполненными значениями метрик. Запрос через `POST value/`.

2. Переведите агента на новое API.  
   Тесты проверяют, что агент экспортирует и обновляет на сервере метрики, описанные в первом инкременте.
