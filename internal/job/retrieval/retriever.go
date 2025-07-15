package retrieval

import (
	"context"
	"fmt"

	"github.com/y7ut/potami/internal/document"
	"github.com/y7ut/potami/internal/task"
)

var _ task.Job = (*Retriever)(nil)

// Retrieval 检索器Interface
type Retrieval interface {
	Query(ctx context.Context, query string) (document.DocumentCollection, error)
}

type Retriever struct {
	task.JobHelper
	Retrieval Retrieval

	QueryField  string
	OutputField string
}

func (r *Retriever) Handle(ctx context.Context) (err error) {
	defer func() {
		if recoverError := recover(); recoverError != nil {
			err = fmt.Errorf("retrieval panic[%v]", recoverError)
		}
		if err != nil {
			r.Logger().WithError(err).Error("retrieval failed")
			r.SetError(err)
		}
	}()

	docs, err := r.search(ctx)
	if err != nil {
		return
	}

	var compressSize int
	compressSize, err = task.BindWithOption(r, "block_size", 1000)
	if err != nil {
		return
	}

	output := docs.Compress(
		document.WithDepth(task.MustBindWithOption(r, "depth_mode", "depth") == "depth"),
		document.WithSize(compressSize),
	)
	r.SetAttribute(r.OutputField, output)

	r.Logger().WithField(r.OutputField, output).Debug("retrieval complete")
	return
}

// searchFromCorpus use retriever to query
func (r *Retriever) search(ctx context.Context) (document.DocumentCollection, error) {

	query, ok := r.GetAttribute(r.QueryField)
	if !ok {
		err := fmt.Errorf("query field %s not found", r.QueryField)
		return nil, err
	}

	question, ok := query.(string)
	if !ok {
		err := fmt.Errorf("query field %s not string", r.QueryField)
		return nil, err
	}

	docs, err := r.Retrieval.Query(ctx, question)
	if err != nil {
		return nil, err
	}
	return docs, nil
}
