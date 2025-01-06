package search

import (
	"context"
	"fmt"

	"github.com/y7ut/potami/internal/conf"
	"github.com/y7ut/potami/internal/job/retrieval"
	"github.com/y7ut/potami/internal/task"
	"github.com/y7ut/potami/pkg/param"
	"google.golang.org/api/customsearch/v1"
	googleOption "google.golang.org/api/option"
)

var _ retrieval.SearchEngine = (*GoogleCustomSearch)(nil)

type GoogleCustomSearch struct {
	APIKey string
	CX     string

	Debug      bool
	MaxResults int

	options task.WithOption
}

func NewGoogleCustomSearch(options task.WithOption) *GoogleCustomSearch {
	return &GoogleCustomSearch{
		APIKey: conf.GoogleCustomSearch.APIKey,
		CX:     conf.GoogleCustomSearch.CX,

		options: options,
	}
}

// Search Implements SearchEngine
func (gcs *GoogleCustomSearch) Search(ctx context.Context, query string) (retrieval.DocumentCollection, error) {

	if err := gcs.applyParams(); err != nil {
		return nil, err
	}

	if gcs.Debug {
		fmt.Printf("google custom search api input: %s\n", query)
	}
	svc, err := customsearch.NewService(ctx, googleOption.WithAPIKey(gcs.APIKey))
	if err != nil {
		return nil, err
	}
	documents := make([]retrieval.Document, 0)
	if gcs.MaxResults > 10 {
		limit := 10
		for i, page := 1, 1; page <= gcs.MaxResults/limit; i, page = i+10, page+1 {
			currentLimit := limit
			if page == gcs.MaxResults/limit {
				currentLimit = gcs.MaxResults - page*limit
			}
			resp, err := svc.Cse.List().Cx(gcs.CX).Q(query).Start(int64(i)).Num(int64(currentLimit)).Do()
			if err != nil {
				return nil, err
			}
			for _, result := range resp.Items {
				documents = append(documents, retrieval.Document{
					Text:   result.Snippet,
					Name:   result.Title,
					Source: result.Link,
				})
			}
		}

	} else {
		resp, err := svc.Cse.List().Cx(gcs.CX).Q(query).Num(int64(gcs.MaxResults)).Do()
		if err != nil {
			return nil, err
		}
		for _, result := range resp.Items {
			documents = append(documents, retrieval.Document{
				Text:   result.Snippet,
				Name:   result.Title,
				Source: result.Link,
			})
		}

	}

	return documents, nil

}

// applyParams
func (gcs *GoogleCustomSearch) applyParams() error {

	if err := param.Assign(&gcs.Debug, gcs.options.GetOptionWithDefault("debug", false)); err != nil {
		return err
	}
	if err := param.Assign(&gcs.MaxResults, gcs.options.GetOptionWithDefault("limit", 10)); err != nil {
		return err
	}
	if gcs.MaxResults > 100 {
		return fmt.Errorf("google custom search max results error: %d", gcs.MaxResults)
	}
	return nil
}
