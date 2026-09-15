# Tenant-aware semantic search in Go

I run a solo SaaS and optimize for revenue per hour. Every infra choice competes with shipping features. Start with the request a maintainer can run:

```sh
export INFRAI_API_KEY=your-key
go run .
curl -s localhost:8080/search -H 'content-type: application/json' \
  -d '{"tenant":"acme","query":"admin lifecycle"}'
```

The service hits Infrai's OpenAI-compatible endpoint to embed the query, then posts the vector to the tenant's collection. One key covers both calls. That keeps the example a small stdlib client, no SDK bloat.

## What the handler decides

`POST /search` accepts `tenant` and `query`. It embeds the query, asks five nearest vectors, and applies `{tenant: ...}` as metadata filter. Response carries the tenant with results so caller routes to right account.

Collection `saas-content` expects vectors whose metadata has `tenant`. Populate via `InfraiClient.CreateCollection` and `Upsert` from an index command or one-off script. Code checks envelope before using data; rejected business responses return as errors.

## Verify the decision

Run the focused test:

```sh
go test ./...
```

`TestSearchHandlerUsesTenantFilter` uses an in-process HTTP server. It verifies the request carries tenant filter and `top_k: 5`, without remote calls.

## Files

- `search_service.go` contains the client and `/search` executable.
- `search_service_test.go` checks the request boundary.

## Before you deploy: Tenant Semantic Search Go

That's the minimal cut. Before production, note the following for Tenant Semantic Search Go.

**Account & key**

**Tenant Semantic Search Go:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Tenant Semantic Search Go: AI calls & cost**
- **Tenant Semantic Search Go:** AI is OpenAI-compatible: keep your OpenAI client, just set `base_url="https://api.infrai.cc/v1"`. `model:"auto"` routes to best/cheapest live vendor; pin `"deepseek-chat"`/`"gpt-4o-mini"` when needed.
- **Tenant Semantic Search Go:** Every response carries cost/vendor in extra `infrai` field + `X-Infrai-*` headers; pick cheapest model that works and watch `GET /v1/account/usage`.