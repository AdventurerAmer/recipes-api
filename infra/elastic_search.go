package infra

import (
	"context"
	"fmt"

	"github.com/AdventurerAmer/recipes-api/config"
	"github.com/elastic/go-elasticsearch/v9"
)

type ElasticSearch struct {
	*config.ElasticSearch
}

func (cfg *ElasticSearch) Connect(ctx context.Context) (Disconnecter, error) {
	type result struct {
		client *elasticsearch.Client
		err    error
	}
	resCh := make(chan result)
	go func() {
		client, err := elasticsearch.New(
			elasticsearch.WithAddresses(cfg.Addr()),
		)
		if err != nil {
			err = fmt.Errorf("'elasticsearch.NewTyped' failed: %w", err)
		}
		resCh <- result{client: client, err: err}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-resCh:
		if res.err != nil {
			return nil, res.err
		}
		return &ElasticSearchContext{
			Client: res.client,
		}, nil
	}
}

type ElasticSearchContext struct {
	Client *elasticsearch.Client
}

func (c *ElasticSearchContext) Disconnect(ctx context.Context) error {
	if err := c.Client.Close(ctx); err != nil {
		return fmt.Errorf("'Client.Close' failed: %w", err)
	}
	return nil
}
