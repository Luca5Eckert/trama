# trama

Base de API Go organizada como monólito modular. Cada módulo concentra seu fluxo e expõe somente a composição em `internal/<modulo>/module.go`.

## Executar

```bash
go run ./cmd/api
```

Variáveis opcionais: `HTTP_ADDRESS` (padrão `:8080`) e `LOG_LEVEL` (padrão `info`).

## Fluxo de referência: users

`HTTP handler -> application service -> domain -> repository interface -> infrastructure`

- `POST /v1/users` com `{"name":"Ada Lovelace"}` cria um usuário.
- `GET /v1/users/{id}` consulta um usuário.
- `GET /health` verifica a saúde do processo.

O repositório atual é em memória. Para incluir Postgres, adicione um adapter em `internal/users/infrastructure/postgres` que implemente `domain.Repository` e injete-o em `users.NewModule()`.
