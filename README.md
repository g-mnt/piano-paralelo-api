# piano-paralelo-api

Backend da plataforma Piano Paralelo — API REST em Go.

## Stack

- Go 1.22
- Gin v1 (HTTP)
- pgx/v5 (PostgreSQL, sem ORM)
- golang-jwt/jwt/v5
- golang-migrate
- bcrypt custo 12

## Rodar localmente

```bash
# 1. Copie o .env
cp .env.example .env

# 2. Suba o banco
docker compose up -d db

# 3. Instale as dependências Go
go mod tidy

# 4. Rode o servidor (migrations e seed rodam automaticamente)
make run

# Ou tudo junto com docker
docker compose up
```

## Endpoints

### Público

| Método | Path             | Descrição         |
|--------|------------------|-------------------|
| GET    | /health          | Health check      |
| POST   | /auth/register   | Criar conta       |
| POST   | /auth/login      | Login             |

### Autenticado (Bearer token)

| Método | Path                            | Descrição                        |
|--------|---------------------------------|----------------------------------|
| GET    | /profile                        | Perfil do aluno                  |
| PUT    | /profile                        | Atualizar perfil                 |
| GET    | /curriculum/weeks               | Listar semanas do currículo      |
| GET    | /curriculum/weeks/:n            | Detalhes da semana (com tarefas) |
| POST   | /sessions                       | Criar/obter sessão de prática    |
| PATCH  | /sessions/:id/tasks/:taskId     | Marcar/desmarcar tarefa          |
| GET    | /sessions/streak                | Streak atual do aluno            |
| GET    | /repertoire                     | Listar peças com progresso       |
| PATCH  | /repertoire/:id/progress        | Atualizar status de uma peça     |

## Variáveis de ambiente

```env
DATABASE_URL=postgres://piano:piano@localhost:5432/piano_paralelo?sslmode=disable
JWT_SECRET=your-super-secret-key-change-in-production
PORT=8080
ENV=development
```

## Migrations

Rodam automaticamente na startup. Para rodar manualmente:

```bash
make migrate-up
make migrate-down
```
