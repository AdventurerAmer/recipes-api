package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/esapi"
)

type textSearch struct {
	client *elasticsearch.Client
}

func New(client *elasticsearch.Client) (ports.TextSearch, error) {
	mapping := `{
        "mappings": {
            "properties": {
                "id": { "type": "keyword" },
                "createdAt": { "type": "date" }
				"userID": { "type": "keyword" },
				"name": {
                    "type": "text",
                    "analyzer": "standard",
                    "fields": {
                        "keyword": { "type": "keyword" },
                        "suggest": { "type": "completion" }
                    }
                },
                "tags": { "type": "keyword" },
                "ingredients": { "type": "keyword" },
				"instructions": { "type": "keyword" },
				"image": { "type": "keyword" },
				"version": { "type": "long" }
            }
        }
    }`

	req := esapi.IndicesCreateRequest{
		Index: "recipes",
		Body:  strings.NewReader(mapping),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := req.Do(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("create index failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("create index failed: %s", res.String())
	}

	return &textSearch{client: client}, nil
}

func (ts *textSearch) Index(ctx context.Context, index string, id string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("'json.Marshal' failed: %w", err)
	}
	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: id,
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}
	res, err := req.Do(ctx, ts.client)
	if err != nil {
		return fmt.Errorf("indexing error: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("indexing error: %w", err)
	}
	return nil
}

func (ts *textSearch) Search(ctx context.Context, index string, field string, value string, page, perPage int) ([][]byte, int, error) {
	from := (page - 1) * perPage
	query := fmt.Sprintf(`{
		"from": %d,
        "size": %d,
        "query": {
            "match": {
                %q: {
                    "query": %q,
                    "fuzziness": "AUTO"
                }
            }
        }
		"sort": [
            { "createdAt": { "order": "desc" } }
        ]
    }`, from, perPage, field, value)
	res, err := ts.client.Search(
		ts.client.Search.WithIndex(index),
		ts.client.Search.WithBody(strings.NewReader(query)),
	)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()

	var raw map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
		return nil, 0, err
	}

	hitsObj := raw["hits"].(map[string]interface{})
	totalObj := hitsObj["total"].(map[string]interface{})
	total := int(totalObj["value"].(float64))

	pages := total / perPage
	if total%perPage != 0 {
		pages++
	}

	hits := hitsObj["hits"].([]interface{})
	results := make([][]byte, 0, len(hits))
	for _, hit := range hits {
		source := hit.(map[string]any)["_source"]
		data, err := json.Marshal(source)
		if err != nil {
			return nil, 0, err
		}
		results = append(results, data)
	}
	return results, total, nil
}

func (ts *textSearch) Delete(ctx context.Context, index string, id string) error {
	req := esapi.DeleteRequest{
		Index:      index,
		DocumentID: id,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, ts.client)
	if err != nil {
		return fmt.Errorf("error deleting recipe: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == http.StatusNotFound {
			return fmt.Errorf("recipe with ID %s not found", id)
		}
		return fmt.Errorf("delete failed [%s]: %s", res.Status(), res.String())
	}

	return nil
}
