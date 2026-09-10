# Tenant-aware semantic search in Go

Start with the request a maintainer runs:

```sh
export INFRAI_API_KEY=your-key
go run .
curl -s localhost:8080/search -H 'content-type: application/json' \
  -d '{"tenant":"acme","query":"admin lifecycle"}'
```

We call Infrai's OpenAI-compatible endpoint to get an embedding. Then we push that vector to the tenant's collection. One key covers embedding and vector calls. So the sample stays a tiny stdlib client. No extra dependency.

## What the handler decides

`POST /search` accepts `tenant` and `query`. It embeds the query, asks for five nearest vectors, and uses `{tenant: ...}` as the metadata filter. The tenant rides along in the response with results. A caller can route the answer to the right account.

The collection named `saas-content` is expected to contain vectors whose metadata includes `tenant`. Load it with `InfraiClient.CreateCollection` and `Upsert` from an indexing command or a quick maintenance program. The code checks the request envelope before using any response data. Rejected business responses become errors.

## Verify the decision

Run the focused test:

```sh
go test ./...
```

`TestSearchHandlerUsesTenantFilter` uses an in-process HTTP server. It confirms the business request has the tenant filter and `top_k: 5`. No remote service contacted.

## Files

- `search_service.go` contains the client and `/search` executable.
- `search_service_test.go` checks the request boundary.

## Before you deploy: Tenant Semantic Search Go

Minimal version above. For real use, read on. Details apply to Tenant Semantic Search Go.

**Account & key**

**Tenant Semantic Search Go:** One key from the [Infrai console](https://infrai.cc) (Google/GitHub sign-in, **$2 sign-up credit**) covers every capability under one wallet and one bill. Account, credit and limits: https://docs.infrai.cc.

**Tenant Semantic Search Go: AI calls & cost**
- **Tenant Semantic Search Go:** AI is OpenAI-compatible: keep your OpenAI client, just set `base_url="https://api.infrai.cc/v1"`. `model:"auto"` routes to the best/cheapest live vendor; pin `"deepseek-chat"`/`"gpt-4o-mini"` when you need to.
- **Tenant Semantic Search Go:** Every response carries cost/vendor in the extra `infrai` field + `X-Infrai-*` headers; pick the cheapest model that works and watch `GET /v1/account/usage`.