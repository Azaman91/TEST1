❗Прежде чем тестировать проект установите эти библиотеки:❗

🟪 go install github.com/swaggo/swag/cmd/swag@latest

🟨 go get -u github.com/swaggo/http-swagger

🟧 go get github.com/swaggo/http-swagger

🟥 go get github.com/jackc/pgx/v5

🟦 go get github.com/jackc/pgx/v5/pgxpool

🟩 go get -u github.com/golang-migrate/migrate/v4

✅ Создайте таблицу в 001.up.SQL и загрузите в контейнер, запуск go run main.go ✅

⚠️ Проект контейнеризируется только через dockerfile ⚠️

⛔ docker compose я не использовал из за сети отдельной в докер контейнере,я просто не смог подключиться к бд ⛔


 ☢️ СТРУКТУРА ☢️
  
Project >  docs > docs.go - Заглушка 🚫 > swagger.yaml - свагер файл 🟩

Project > migrations > 001.up.SQL - Создание таблицы и начало миграции 🔼 > 002.down.SQL - Откат миграции 🔽

Project > REST > rest.go - Все ручки и операции и все данные обрабатываются в нём ♾️

Project > docker-compose.yaml - Докер контейнер для PostgreSQL ⛔

Project > dockerfile - Настройки контейнера 🐳

Project > go.mod

Project > go.sum

Project > info.env - Файл с нужной информации ⚙️

Project > main.go - Вход в программу и запуск программы 🌐

