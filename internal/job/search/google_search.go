package search

import (
	"context"
	"fmt"

	"github.com/y7ut/potami/internal/conf"
	customsearch "google.golang.org/api/customsearch/v1"
	"google.golang.org/api/option"
)

type GoogleCustomSearch struct {
	APIKey string
	CX     string
	Query  string
	Debug  bool

	MaxResults int
}

func NewGoogleCustomSearch() *GoogleCustomSearch {
	return &GoogleCustomSearch{
		APIKey: conf.GoogleCustomSearch.APIKey,
		CX:     conf.GoogleCustomSearch.CX,
	}
}

func (gcs *GoogleCustomSearch) Search(ctx context.Context, param *SearchParam) ([]Document, error) {

	if err := gcs.applyParams(param); err != nil {
		return nil, err
	}

	if gcs.Debug {
		fmt.Printf("google custom search api input: %s\n", gcs.Query)
	}
	svc, err := customsearch.NewService(ctx, option.WithAPIKey(gcs.APIKey))
	if err != nil {
		return nil, err
	}
	documents := make([]Document, 0)
	if gcs.MaxResults > 10 {
		limit := 10
		for i, page := 1, 1; page <= gcs.MaxResults/limit; i, page = i+10, page+1 {
			currentLimit := limit
			if page == gcs.MaxResults/limit {
				currentLimit = gcs.MaxResults - page*limit
			}
			resp, err := svc.Cse.List().Cx(gcs.CX).Q(gcs.Query).Start(int64(i)).Num(int64(currentLimit)).Do()
			if err != nil {
				return nil, err
			}
			for _, result := range resp.Items {
				documents = append(documents, Document{
					Content: result.Snippet,
					Title:   result.Title,
					URL:     result.Link,
				})
			}
		}

	} else {
		resp, err := svc.Cse.List().Cx(gcs.CX).Q(gcs.Query).Num(int64(gcs.MaxResults)).Do()
		if err != nil {
			return nil, err
		}
		for _, result := range resp.Items {
			documents = append(documents, Document{
				Content: result.Snippet,
				Title:   result.Title,
				URL:     result.Link,
			})
		}

	}

	return documents, nil

}

// applyParams
func (gcs *GoogleCustomSearch) applyParams(param *SearchParam) error {

	gcs.Debug = param.Debug
	if param.Limit > 100 {
		return fmt.Errorf("google custom search max results error: %d", gcs.MaxResults)
	}
	gcs.MaxResults = param.Limit
	gcs.Query = param.Query
	return nil
}
