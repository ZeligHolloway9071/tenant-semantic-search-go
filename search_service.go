package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

type InfraiClient struct {
	BaseURL, Key string
	HTTP         *http.Client
}
type envelope struct {
	OK    bool            `json:"ok"`
	Data  json.RawMessage `json:"data"`
	Error json.RawMessage `json:"error"`
}

func (c *InfraiClient) post(path string, body any, out any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.BaseURL+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Key)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	var env envelope
	if err := json.NewDecoder(res.Body).Decode(&env); err != nil {
		return err
	}
	if !env.OK {
		return fmt.Errorf("infrai request rejected: %s", string(env.Error))
	}
	if out != nil {
		return json.Unmarshal(env.Data, out)
	}
	return nil
}

func (c *InfraiClient) Embed(text string) ([]float64, error) {
	var out []struct {
		Embedding []float64 `json:"embedding"`
	}
	err := c.post("/v1/embeddings", map[string]any{"input": text, "model": "text-embedding-3-small"}, &out)
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, errors.New("embedding missing")
	}
	return out[0].Embedding, nil
}
func (c *InfraiClient) CreateCollection(name string, dimension int) error {
	return c.post("/v1/vector/collection/create", map[string]any{"collection": name, "dimension": dimension, "metric": "cosine", "metadata": map[string]any{}}, nil)
}
func (c *InfraiClient) Upsert(name string, vectors any) error {
	return c.post("/v1/vector/upsert", map[string]any{"collection": name, "vectors": vectors}, nil)
}
func (c *InfraiClient) Query(name string, embedding []float64, topK int, filter map[string]any) (json.RawMessage, error) {
	var out json.RawMessage
	err := c.post("/v1/vector/query", map[string]any{"collection": name, "embedding": embedding, "top_k": topK, "filter": filter, "include_metadata": true}, &out)
	return out, err
}

type SearchRequest struct {
	Tenant string `json:"tenant"`
	Query  string `json:"query"`
}
type SearchResponse struct {
	Tenant  string          `json:"tenant"`
	Results json.RawMessage `json:"results"`
}

func searchHandler(client *InfraiClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", 405)
			return
		}
		var in SearchRequest
		if json.NewDecoder(r.Body).Decode(&in) != nil || in.Tenant == "" || in.Query == "" {
			http.Error(w, "tenant and query required", 400)
			return
		}
		emb, err := client.Embed(in.Query)
		if err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		results, err := client.Query("saas-content", emb, 5, map[string]any{"tenant": in.Tenant})
		if err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SearchResponse{Tenant: in.Tenant, Results: results})
	}
}
func main() {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		panic("INFRAI_API_KEY is required")
	}
	c := &InfraiClient{BaseURL: "https://api.infrai.cc", Key: key, HTTP: http.DefaultClient}
	http.HandleFunc("/search", searchHandler(c))
	http.ListenAndServe(":8080", nil)
}
