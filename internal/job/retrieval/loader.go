package retrieval

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/y7ut/potami/internal/task"
)

type Loader struct {
	ResourceExtractor    Extractor
	ChunkSplitter        Splitter
	KnowledgeBaseBuilder KnowledgeBaseBuilder

	QueryField  string
	OutputField string
	task.JobHelper
}

func (l *Loader) Handle(ctx context.Context) (err error) {
	defer func() {
		if recoverError := recover(); recoverError != nil {
			err = fmt.Errorf("dialog panic [%v]", recoverError)
		}
		if err != nil {
			l.Logger().WithError(err).Error("dialog failed")
			l.SetError(err)
		}
	}()

	resourceAddress, ok := l.GetAttribute(l.QueryField)

	if !ok {
		err := fmt.Errorf("query field %s not found", l.QueryField)
		return err
	}

	address, ok := resourceAddress.(string)
	if !ok {
		err := fmt.Errorf("query field %s not string", l.QueryField)
		return err
	}

	resourceData, err := loadResource(address)
	if err != nil {
		err = fmt.Errorf("load resource error: %v", err)
		return err
	}

	resource, err := l.ResourceExtractor.Extract(ctx, resourceData)
	if err != nil {
		return err
	}

	l.SetAttribute(l.OutputField, resource)

	docs := l.ChunkSplitter.Split(ctx, resource)

	for _, chunk := range docs {
		err = l.KnowledgeBaseBuilder.Store(ctx, chunk)
		if err != nil {
			l.Logger().WithError(err).Warnf("store document error: %v", err)
		}
	}

	return nil
}

type bytesReader struct {
	*bytes.Reader
}

func (r *bytesReader) Close() error {
	return nil
}

// loadFromFile loads resource from file
func loadFromFile(source string) (io.ReadSeekCloser, error) {
	return os.Open(source)
}

// loadFromURL downloads resource from url
func loadFromURL(source string) (io.ReadSeekCloser, error) {
	resp, err := http.Get(source)
	if err != nil {

		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &bytesReader{bytes.NewReader(body)}, nil
}

// loadFromBase64 converts base64 string to io.ReadSeekCloser
func loadFromBase64(data []byte) (io.ReadSeekCloser, error) {
	var base64Buffer []byte
	_, err := base64.StdEncoding.Decode(data, base64Buffer)
	if err != nil {
		err = fmt.Errorf("base64 decode error: %v", err)
		return nil, err
	}
	return &bytesReader{bytes.NewReader(base64Buffer)}, nil
}

func resourceType(source string) string {
	isUrl := func(source string) bool {
		// 简单判定是否为URL，例如以 "http://" 或 "https://" 开头
		return len(source) >= 7 && (source[:7] == "http://" || len(source) >= 8 && source[:8] == "https://")
	}
	isFile := func(source string) bool {
		return len(source) >= 7 && source[:7] == "file://"
	}
	isBase64 := func(source string) bool {
		return len(source) >= 7 && source[:7] == "base64://"
	}
	if isUrl(source) {
		return "url"
	} else if isFile(source) {
		return "file"
	} else if isBase64(source) {
		return "base64"
	} else {
		return "unknown"
	}
}

func loadResource(address string) (io.ReadSeekCloser, error) {
	switch resourceType(address) {
	case "url":
		return loadFromURL(address)
	case "file":
		return loadFromFile(address[7:])
	case "base64":
		return loadFromBase64([]byte(address[9:]))
	default:
		return nil, fmt.Errorf("unknown resource type: %s", address)
	}
}
