package textsearch

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/AdventurerAmer/recipes-api/internal/core/ports"
	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/typedapi/esdsl"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	estypes "github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/refresh"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/sortorder"
)

type textSearch struct {
	client *elasticsearch.TypedClient
}

func NewElasticSearch(client *elasticsearch.TypedClient) (ports.TextSearch, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	index := "recipes"

	exists, err := client.Indices.Exists(index).Do(ctx)
	if err != nil {
		return nil, fmt.Errorf("'client.Indices.Exists' failed: %w", err)
	}

	if !exists {
		mappings := esdsl.NewTypeMapping()
		mappings.AddProperty("id", esdsl.NewKeywordProperty())
		mappings.AddProperty("createdAt", esdsl.NewDateProperty())
		mappings.AddProperty("userId", esdsl.NewKeywordProperty())
		mappings.AddProperty("name", esdsl.NewTextProperty().
			Analyzer("standard").
			Fields(map[string]types.Property{
				"keyword": esdsl.NewKeywordProperty(),
				"suggest": esdsl.NewCompletionProperty(),
			}),
		)
		mappings.AddProperty("tags", esdsl.NewKeywordProperty())
		mappings.AddProperty("ingredients", esdsl.NewKeywordProperty())
		mappings.AddProperty("instructions", esdsl.NewKeywordProperty())
		mappings.AddProperty("imageURL", esdsl.NewKeywordProperty())
		mappings.AddProperty("version", esdsl.NewLongNumberProperty())

		numberOfShards := "1"
		numberOfReplicas := "1"
		settings := &estypes.IndexSettings{
			NumberOfShards:   &numberOfShards,
			NumberOfReplicas: &numberOfReplicas,
		}
		if _, err := client.Indices.Create(index).Mappings(mappings).Settings(settings).Do(ctx); err != nil {
			return nil, fmt.Errorf("'client.Indices.Create' failed: %w", err)
		}
	}

	return &textSearch{client: client}, nil
}

func (ts *textSearch) Index(ctx context.Context, index, id string, document any) error {
	_, err := ts.client.Index(index).
		Id(id).
		Document(document).
		Refresh(refresh.True).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("'client.Index' failed: %w", err)
	}
	return nil
}

func (ts *textSearch) Search(ctx context.Context, index, field, value string, cursor string, limit int64) ([]json.RawMessage, string, error) {
	type elasticCursor struct {
		After []types.FieldValue `json:"after"`
	}

	keepAlive := "1m"
	var ec elasticCursor
	pitRes, err := ts.client.OpenPointInTime(index).
		KeepAlive(keepAlive).
		Do(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("'ts.client.OpenPointInTime' failed: %w", err)
	}

	if cursor != "" {
		data, err := hex.DecodeString(cursor)
		if err != nil {
			return nil, "", fmt.Errorf("'hex.DecodeString' failed: %w", err)
		}
		if err := json.Unmarshal(data, &ec); err != nil {
			return nil, "", fmt.Errorf("'json.Unmarshal' failed: %w", err)
		}
	}

	req := ts.client.Search().
		Size(int(limit+1)).
		Query(esdsl.NewMatchPhrasePrefixQuery(field, value)).
		Sort(esdsl.NewSortOptions().
			AddSortOption("createdAt", esdsl.NewFieldSort(sortorder.Desc)),
			esdsl.NewSortOptions().
				AddSortOption("_shard_doc", esdsl.NewFieldSort(sortorder.Asc))).
		Pit(&types.PointInTimeReference{
			Id:        pitRes.Id,
			KeepAlive: keepAlive,
		})
	if len(ec.After) > 0 {
		req = req.SearchAfterValues(ec.After)
	}

	res, err := req.Do(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("'req.Do' failed: %w", err)
	}

	hits := res.Hits.Hits
	if len(hits) == 0 {
		return nil, "", nil
	}

	results := make([]json.RawMessage, 0, len(hits))

	for _, hit := range hits {
		results = append(results, hit.Source_)
	}

	if len(hits) == int(limit+1) {
		lastHit := hits[len(hits)-2]
		ec.After = lastHit.Sort

		data, err := json.Marshal(ec)
		if err != nil {
			return nil, "", fmt.Errorf("'json.Marshal' failed: %w", err)
		}

		ret := results[:limit]
		return ret, hex.EncodeToString(data), nil
	}
	return results, "", nil
}

func (ts *textSearch) Delete(ctx context.Context, index string, id string) error {
	if _, err := ts.client.Delete(index, id).Do(ctx); err != nil {
		return fmt.Errorf("'client.Delete' failed: %w", err)
	}
	return nil
}
