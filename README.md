# trama

Base de API Go organizada como monólito modular. Cada módulo concentra seu fluxo e expõe somente a composição em `internal/<modulo>/module.go`.

## Executar

```bash
docker compose up -d
go run ./cmd/api
```

Variáveis opcionais: `HTTP_ADDRESS` (padrão `:8080`), `LOG_LEVEL` (padrão `info`) e `DATABASE_URL` (padrão `postgres://postgres:postgres@localhost:5432/trama?sslmode=disable`).

## Arquitetura dos módulos

A direção de dependências é:

```text
Presentation -> Application -> Domain <- Infrastructure
```

- `presentation`: controller HTTP, DTOs e mapeamento da borda;
- `application`: use cases separados em `command` e `query`;
- `domain`: modelos, erros e portas; não conhece HTTP, JSON, pgx ou adapters;
- `infrastructure`: implementações das portas, como PostgreSQL, memória, clock e geração de ID;
- `module.go`: composition root e único ponto do módulo autorizado a conhecer as implementações concretas usadas no wiring.

O teste em `internal/users/architecture` protege as importações proibidas entre camadas.

## Fluxo de referência: users

- `POST /v1/users` com `{"name":"Ada Lovelace"}` cria um usuário;
- `GET /v1/users/{id}` consulta um usuário;
- `GET /v1/users` lista usuários;
- `GET /health` verifica a saúde do processo.

A presentation usa DTOs próprios. Modelos de domínio não possuem tags JSON e não são serializados diretamente.

Erros HTTP seguem envelope estável:

```json
{
  "error": {
    "code": "invalid_name",
    "message": "name is required"
  }
}
```

## Banco de dados

`database.Open` cuida apenas da conexão e do lifecycle do pool. `database.Migrate` aplica o schema bootstrap atual de forma explícita. Antes de adicionar tabelas do módulo `production`, o projeto deve adotar migrations versionadas; não será criado um framework próprio de migrations para isso.
