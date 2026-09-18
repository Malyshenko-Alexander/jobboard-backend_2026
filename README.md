# Бэк сайта для поиска работы

## Сервисы

| Сервис | Порт | Назначение |
|--------|------|------------|
| auth-service | 8081 | Регистрация, вход, JWT |
| applicant-service | 8082 | Кабинет соискателя (профиль, резюме) |
| employer-service | 8083 | Кабинет работодателя, публичные данные компании |
| vacancy-service | 8084 | Поиск и CRUD вакансий |
| PostgreSQL | 5432 | 4 отдельные БД |
| RabbitMQ | 5672 / 15672 | Асинхронные события |

## Перед стартом

```bash
cp .env.example .env

docker compose up --build -d
```

Swagger-ы:

- Auth Swagger: http://localhost:8081/swagger/index.html
- Applicant: http://localhost:8082/swagger/index.html
- Employer: http://localhost:8083/swagger/index.html
- Vacancy: http://localhost:8084/swagger/index.html
- RabbitMQ UI: http://localhost:15672 (логин/пароль из `.env`)

Стоп:

```bash
docker compose down
```

Очистка БД:

```bash
docker compose down -v
```

## Документация

- [Ручное развертывание на Linux](DEPLOY.md)
- CI: `.github/workflows/ci.yml`
