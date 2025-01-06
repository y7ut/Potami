package retrieval

import (
	"context"
	"fmt"

	"github.com/y7ut/potami/internal/task"
	"github.com/y7ut/potami/pkg/param"
)

// Retriever 检索器Interface
type Retriever interface {
	Query(ctx context.Context, query string) (DocumentCollection, error)
}

type Retrieval struct {
	task.JobHelper
	Retriever Retriever

	QueryField  string
	OutputField string
}

func (r *Retrieval) Handle(ctx context.Context) (err error) {
	defer func() {
		if recoverError := recover(); recoverError != nil {
			err = fmt.Errorf("retrieval panic[%v]", recoverError)
		}
		if err != nil {
			r.Logger().WithError(err).Error("retrieval failed")
			r.SetError(err)
		}
	}()

	docs, err := r.query(ctx)
	if err != nil {
		return
	}

	defaultCompressMethod := func(doc Document) string {
		return fmt.Sprintf("《%s》\n%s\n", doc.Name, doc.Text)
	}

	var compressSize int
	if paramError := param.Assign(&compressSize, r.GetOptionWithDefault("block_size", 1000)); paramError != nil {
		err = fmt.Errorf("search block size type error, error: %v", paramError)
		return
	}
	var depthMode string
	if paramError := param.Assign(&depthMode, r.GetOptionWithDefault("depth_mode", "depth")); paramError != nil {
		err = fmt.Errorf("search depth mode type error, error: %v", paramError)
		return
	}

	output := docs.Compress(defaultCompressMethod, compressSize, depthMode == "depth")
	r.SetAttribute(r.OutputField, output)

	r.Logger().WithField(r.OutputField, output).Debug("retrieval complete")
	return
}

// query use retriever to query
func (r *Retrieval) query(ctx context.Context) (DocumentCollection, error) {

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

	docs, err := r.Retriever.Query(ctx, question)
	if err != nil {
		return nil, err
	}
	return docs, nil
}
