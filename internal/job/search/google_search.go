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

func (gcs *GoogleCustomSearch) Search(ctx context.Context, query string, options ...SearchEngineOption) ([]Document, error) {

	if err := gcs.applyOptions(options...); err != nil {
		return nil, err
	}
	gcs.Query = query
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

func (gcs *GoogleCustomSearch) applyOptions(options ...SearchEngineOption) error {
	om := DefaultOptionsManager()
	for _, applyOption := range options {
		applyOption(om)
	}
	gcs.Debug = om.Debug
	if gcs.MaxResults > 100 {
		gcs.MaxResults = 100
	}
	gcs.MaxResults = om.Limit

	return nil
}
