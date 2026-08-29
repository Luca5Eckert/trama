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

Os testes em `internal/<modulo>/architecture` protegem as importações proibidas entre camadas.

## Users

- `POST /v1/users` com `{"name":"Ada Lovelace"}` cria um usuário;
- `GET /v1/users/{id}` consulta um usuário;
- `GET /v1/users` lista usuários;
- `GET /health` verifica a saúde do processo.

A presentation usa DTOs próprios. Modelos de domínio não possuem tags JSON e não são serializados diretamente.

## Production — sequência de tamanhos

A sequência de tamanhos é uma configuração singleton do processo produtivo:

- `GET /v1/production/size-sequence` consulta a configuração vigente;
- `PUT /v1/production/size-sequence` substitui a configuração inteira.

A ordem é definida exclusivamente por `position`; nomes não são ordenados alfabeticamente e nenhuma sequência como `P -> M -> G` é assumida pelo código.

Exemplo:

```json
{
  "items": [
    {"name": "P", "position": 10},
    {"name": "M", "position": 20}
  ]
}
```

O `PUT` é idempotente: repetir o mesmo estado lógico não duplica itens nem altera `updatedAt`.

## Erros HTTP

As bordas HTTP usam envelope estável:

```json
{
  "error": {
    "code": "invalid_size_sequence",
    "message": "invalid size sequence"
  }
}
```

## Banco de dados

`database.Open` cuida somente da conexão e do lifecycle do pool. `database.Migrate` executa migrations SQL versionadas e embutidas em `internal/platform/database/migrations` usando Tern.

A tabela de controle é `public.trama_schema_version`. O schema de `production` é criado por migration; SQL e transações de persistência permanecem escondidos nos adapters PostgreSQL.
