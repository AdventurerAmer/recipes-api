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

func (ts *textSearch) Search(ctx context.Context, index, cursor string, limit int64) ([]json.RawMessage, string, error) {
	type elasticCursor struct {
		PitId string             `json:"pitId"`
		Next  []types.FieldValue `json:"next"`
	}
	keepAlive := "2m"
	var ec elasticCursor
	if cursor == "" {
		pitRes, err := ts.client.OpenPointInTime(index).
			KeepAlive(keepAlive).
			Do(ctx)
		if err != nil {
			return nil, "", fmt.Errorf("open PIT: %w", err)
		}
		ec.PitId = pitRes.Id
	} else {
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
		Sort(esdsl.NewSortOptions().AddSortOption("createdAt", esdsl.NewFieldSort(sortorder.Desc)),
			esdsl.NewSortOptions().AddSortOption("_shard_doc", esdsl.NewFieldSort(sortorder.Asc))).
		Pit(&types.PointInTimeReference{
			Id:        ec.PitId,
			KeepAlive: keepAlive,
		})
	if len(ec.Next) > 0 {
		req = req.SearchAfterValues(ec.Next)
	}
	res, err := req.Do(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("'req.Do' failed: %w", err)
	}
	// Elasticsearch may return an updated PIT ID – always use the latest one
	if res.PitId != nil {
		ec.PitId = *res.PitId
	}

	hits := res.Hits.Hits
	if len(hits) == 0 {
		return nil, "", nil
	}

	results := make([]json.RawMessage, len(hits))

	for _, hit := range hits {
		results = append(results, hit.Source_)
	}

	if len(hits) == int(limit)+1 {
		lastHit := hits[len(hits)-1]
		ec.Next = lastHit.Sort

		data, err := json.Marshal(ec)
		if err != nil {
			return nil, "", fmt.Errorf("'json.Marshal' failed: %w", err)
		}

		return results[:limit], hex.EncodeToString(data), nil
	}
	return results, "", nil
}

func (ts *textSearch) Delete(ctx context.Context, index string, id string) error {
	if _, err := ts.client.Delete(index, id).Do(ctx); err != nil {
		return fmt.Errorf("'client.Delete' failed: %w", err)
	}
	return nil
}
