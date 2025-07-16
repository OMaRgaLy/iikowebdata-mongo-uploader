# iikowebdata-mongo-updater

Утилита `iikowebdata-mongo-updater` загружает данные "iiko web" из заранее сгенерированного JSON в ваши документы MongoDB по URI

## 📂 Структура проекта

В корне проекта:
```
├── main.go        # точка входа, вся логика обновления
├── go.mod
└── go.sum
```

> JSON-файл с данными (`restaurants_with_iikoweb_data.json`) должен находиться рядом с `main.go`.

## ⚙️ Требования

- Go 1.20+
- Модуль MongoDB-драйвера: `go.mongodb.org/mongo-driver`

## Установка

```bash
git clone https://github.com/OMaRgaLy/iikowebdata-mongo-uploader.git
cd iikowebdata-mongo-updater
```

## 🛠️ Сборка

```bash
go mod tidy
go build -o iikowebdata-mongo-updater main.go
```

Получится исполняемый файл `iikowebdata-mongo-updater` в текущей директории.

## 🚀 Запуск

1. Положите рядом с бинарником файл  
   ```
   restaurants_with_iikoweb_data.json
   ```
2. Отредактируйте в `main.go` строку подключения, если нужно:
   ```go
   uri := "mongodb+srv://<USER>:<PASS>@your-cluster.mongodb.net/?authMechanism=MONGODB-AWS&authSource=%24external"
   ```
3. Запустите утилиту:
   ```bash
   ./iikowebdata-mongo-updater
   ```
4. Когда программа спросит:
   ```
   Введите имя базы данных:
   ```
   — введите, например, `Kwaaka`  
   ```
   Введите имя коллекции:
   ```
   — введите, например, `restaurants`

Утилита:
- Открывает `restaurants_with_iikoweb_data.json`  
- Подключается по `uri`  
- Для каждого `restaurant_id` (конвертируется в ObjectID) проверяет текущие поля  
- Обновляет вложенный документ `iiko_cloud` только если данные отличаются  
- В конце выводит статистику: сколько обновлено, сколько пропущено

## 📑 Формат JSON

```json
{
  "restaurants_with_iikoweb_data": [
    {
      "restaurant_id": "123asd456fgh7j8k9l",
      "iiko_web_domain": "rest-domain.iikoweb.ru",
      "iiko_web_login": "login",
      "iiko_web_password": "password"
    }
    // …
  ]
}
```

## 📝 Лицензия
MIT License 
