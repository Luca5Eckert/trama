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
    {"name": "2", "position": 10},
    {"name": "4", "position": 20},
    {"name": "6", "position": 30}
  ]
}
```

O `PUT` é idempotente: repetir o mesmo estado lógico não duplica itens nem altera `updatedAt`.

## Production — recebimento

`POST /v1/entries` registra uma entrada pelas cores presentes. Quantidade de peças não é obrigatória.

```json
{
  "colors": ["Preto", "Azul", "Bege"]
}
```

Cada cor cria exatamente um `ColorBatch`, preservando a ordem do request. O lote começa em `WAITING` e recebe `SizeRun` em `PENDING` para cada item da sequência vigente. Os tamanhos são copiados para o lote como snapshot; alterar a configuração depois não modifica entradas já recebidas.

A criação da entrada, dos lotes e dos size runs ocorre em uma única transação PostgreSQL. A API responde `201 Created` e `Location: /v1/entries/{id}`.

## Production — consultas e fila

As leituras operacionais são expostas sem carregar o agregado de escrita para montar listagens:

- `GET /v1/entries/{id}` detalha uma entrada e seus lotes;
- `GET /v1/entries?limit=50&offset=0` lista entradas em `receivedAt DESC`;
- `GET /v1/color-batches/{id}` detalha um lote e seus size runs;
- `GET /v1/color-batches?status=WAITING&entryId=<id>&limit=50&offset=0` consulta a fila.

`limit` usa 50 por padrão, aceita de 1 a 100, e `offset` começa em 0. Listas vazias retornam `200` com `items: []`; recursos individuais inexistentes retornam `404`.

A ordem canônica da fila é `created_at ASC, entry_id ASC, position ASC, id ASC`. Isso é somente uma ordenação técnica determinística; ainda não existe regra global de prioridade entre cores de entradas diferentes.

`currentSize` e `nextSize` são projeções dos `SizeRun` persistidos e não colunas duplicadas:

- `WAITING`: `currentSize = null` e `nextSize = primeiro PENDING`;
- `IN_PRODUCTION`: `currentSize = primeiro IN_PROGRESS` e `nextSize = primeiro PENDING`;
- `COMPLETED`: ambos `null`.

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
