`# Для запуска

1. Создать файл в папке проекта (redditclone) `.env` с содержимым:
```env
MYSQL_DSN="root:love@tcp(mysql:3306)/golang?charset=utf8&interpolateParams=true"
MONGO_URL="mongodb://mongodb"
MONGO_DB="golang"
MONGO_COLLECTION="posts"
TOKEN_SECRET="supersecret"
MIGRATION_DIR="./06_databases/99_hw/redditclone/migrations/_sql"
TEMPLATE_DIR="./06_databases/99_hw/redditclone/static/html/index.html"
STATIC_DIR="./06_databases/99_hw/redditclone/static"
```
2. Сделать файл исполняемым: `chmod +x ./start`
3. `./start.sh`