package infra

import (
	"context"
	"fmt"

	"github.com/elastic/go-elasticsearch/v9"
)

type ElasticSearchConfig struct {
	Address string `cfg:"address"`
}

type ElasticSearchContext struct {
	Client *elasticsearch.TypedClient
}

func connectToElasticSearch(ctx context.Context, cfg ElasticSearchConfig) (ElasticSearchContext, error) {
	type result struct {
		client *elasticsearch.TypedClient
		err    error
	}
	resCh := make(chan result)
	go func() {
		client, err := elasticsearch.NewTyped(
			elasticsearch.WithAddresses(cfg.Address),
		)
		if err != nil {
			err = fmt.Errorf("'elasticsearch.NewTyped' failed: %w", err)
		}
		resCh <- result{client: client, err: err}
	}()
	select {
	case <-ctx.Done():
		return ElasticSearchContext{}, ctx.Err()
	case res := <-resCh:
		if res.err != nil {
			return ElasticSearchContext{}, res.err
		}
		return ElasticSearchContext{
			Client: res.client,
		}, nil
	}
}

func disconnectFromElasticSearch(ctx context.Context, elasticSearchCtx ElasticSearchContext) error {
	if err := elasticSearchCtx.Client.Close(ctx); err != nil {
		return fmt.Errorf("'Client.Close' failed: %w", err)
	}
	return nil
}
