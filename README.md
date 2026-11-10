# ITLib

## Инструкция по запуску проекта

### Клониорование репозитория
```bash
git clone https://github.com/penkkk/16.20_team_mkrtchyan.git
cd 16.20_team_mkrtchyan
```

### Скопировать шаблоны переменных окружения
```bash
cp backend/.env.example backend/.env
cp infra/garage/.env.example infra/garage/.env
cp infra/garage/garage.toml.example infra/garage/garage.toml
cp infra/garage/webui/.env.example infra/garage/webui/.env
```
Заполнить созданные файлы, включая токены в garage.toml и хеш пароля в webui/.env

### Запустить сборку и поднять контейнеры
```bash
docker compose --env-file backend/.env config -q
docker compose --env-file backend/.env up -d --build
```
### Проверка
```bash
curl http://whoami.localhost
```
