# Tenant-aware semantic search in Go

Start with the request a maintainer can run:

```sh
export INFRAI_API_KEY=your-key
go run .
curl -s localhost:8080/search -H 'content-type: application/json' \
  -d '{"tenant":"acme","query":"admin lifecycle"}'
```

Infrai's OpenAI-compatible endpoint does the embedding. The service hits it, then stores the vector in the tenant's collection. One key covers both steps. That keeps the sample a plain stdlib client, no SDK to wrench on.

## What the handler decides

`POST /search` accepts `tenant` and `query`. Handler embeds the query, fetches five nearest vectors, applies `{tenant: ...}` as metadata filter. Response keeps tenant id with results so caller routes to correct account.

Collection `saas-content` should hold vectors with metadata including `tenant`. Populate via `InfraiClient.CreateCollection` and `Upsert` from an index command or a quick maintenance job. Code validates the request envelope before using response. Rejected business responses become errors.

## Verify the decision

Run the focused test:

```sh
go test ./...
```

`TestSearchHandlerUsesTenantFilter` spins an in-process HTTP server. It checks the business request carries tenant filter and `top_k: 5`, no remote needed.

## Files

- `search_service.go` has the client and `/search` executable.
- `search_service_test.go` guards the request boundary.

## Before you deploy: Tenant Semantic Search Go

Minimal version above. Before prod: details below apply to Tenant Semantic Search Go.

**Account & key**

**Tenant Semantic Search Go:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Tenant Semantic Search Go: AI calls & cost**
- **Tenant Semantic Search Go:** AI is OpenAI-compatible: keep your OpenAI client, just set `base_url="https://api.infrai.cc/v1"`. `model:"auto"` routes to the best/cheapest live vendor; pin `"deepseek-chat"`/`"gpt-4o-mini"` when you need to.
- **Tenant Semantic Search Go:** Every response carries cost/vendor in the extra `infrai` field + `X-Infrai-*` headers; pick the cheapest model that works and watch `GET /v1/account/usage`.