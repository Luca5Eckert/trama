# trama

Base de API Go organizada como monólito modular. Cada módulo concentra seu fluxo e expõe somente a composição em `internal/<modulo>/module.go`.

## Executar

```bash
go run ./cmd/api
```

Inicie o PostgreSQL local:

```bash
docker compose up -d
```

Depois execute a API. O schema é criado automaticamente na inicialização.

Variáveis opcionais: `HTTP_ADDRESS` (padrão `:8080`), `LOG_LEVEL` (padrão `info`) e
`DATABASE_URL` (padrão `postgres://postgres:postgres@localhost:5432/trama?sslmode=disable`).

## Fluxo de referência: users

`HTTP handler -> application service -> domain -> repository interface -> infrastructure`

- `POST /v1/users` com `{"name":"Ada Lovelace"}` cria um usuário.
- `GET /v1/users/{id}` consulta um usuário.
- `GET /health` verifica a saúde do processo.

Os usuários são persistidos em PostgreSQL. O adapter em memória permanece disponível somente para testes unitários.
