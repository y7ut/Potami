package search

import (
	"context"
	"fmt"

	"github.com/y7ut/potami/internal/conf"
	"github.com/y7ut/potami/internal/document"
	"github.com/y7ut/potami/internal/task"
	"google.golang.org/api/customsearch/v1"
	googleOption "google.golang.org/api/option"
)

var _ SearchEngine = (*GoogleCustomSearch)(nil)

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
func (gcs *GoogleCustomSearch) Search(ctx context.Context, query string) (document.DocumentCollection, error) {

	if task.MustBindWithOption(gcs.options, "debug", false) {
		fmt.Printf("google custom search api input: %s\n", query)
	}

	maxResults := task.MustBindWithOption(gcs.options, "limit", 10)
	if maxResults > 100 {
		return nil, fmt.Errorf("google custom search max results error: %d", maxResults)
	}

	svc, err := customsearch.NewService(ctx, googleOption.WithAPIKey(gcs.APIKey))
	if err != nil {
		return nil, err
	}
	documents := make([]*document.Document, 0)

	if maxResults > 10 {

		limit := 10
		for i, page := 1, 1; page <= maxResults/limit; i, page = i+10, page+1 {
			currentLimit := limit
			if page == maxResults/limit {
				currentLimit = maxResults - page*limit
			}
			resp, err := svc.Cse.List().Cx(gcs.CX).Q(query).Start(int64(i)).Num(int64(currentLimit)).Do()
			if err != nil {
				return nil, err
			}
			for _, result := range resp.Items {
				documents = append(documents, &document.Document{
					Text: result.Snippet,
					Name: result.Title,
					Source: &document.Resource{
						Name:     result.Title,
						Address:  result.Link,
						MineType: "text/html",
					},
				})
			}
		}

	} else {
		resp, err := svc.Cse.List().Cx(gcs.CX).Q(query).Num(int64(maxResults)).Do()
		if err != nil {
			return nil, err
		}
		for _, result := range resp.Items {
			documents = append(documents, &document.Document{
				Text: result.Snippet,
				Name: result.Title,
				Source: &document.Resource{
					Name:     result.Title,
					Address:  result.Link,
					MineType: "text/html",
				},
			})
		}

	}

	return documents, nil

}
