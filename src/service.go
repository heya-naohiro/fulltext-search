package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/blugelabs/bluge"
	"github.com/ikawaha/blugeplugin/analysis/lang/ja"
)

type Service struct {
	directoryPath string
	config        bluge.Config
	writer        *bluge.Writer
}

type QueryResult struct {
	filepath string
}

func (s QueryResult) String() string {
	return s.filepath
}

func NewService(directory string) (*Service, error) {
	if f, err := os.Stat(directory); os.IsNotExist(err) || !f.IsDir() {
		return nil, fmt.Errorf("ServiceCreateErr: %s", err)
	}
	indexPath := filepath.Join(directory, ".blugeindex")
	config := bluge.DefaultConfig(indexPath)
	writer, err := bluge.OpenWriter(config)
	if err != nil {
		return nil, fmt.Errorf("ServiceCreateErr: %s", err)
	}

	return &Service{
		directoryPath: directory,
		config:        config,
		writer:        writer,
	}, nil

}

func (s Service) CreateIndex() error {

	err := filepath.Walk(s.directoryPath, func(p string, info os.FileInfo, err error) error {
		if !info.IsDir() && filepath.Ext(p) == ".txt" {
			// path = id
			// text file only now...
			fmt.Println(p)
			fp, err := os.Open(p)
			if err != nil {
				return err
			}
			defer fp.Close()
			data, err := ioutil.ReadAll(fp)
			if err != nil {
				return err
			}

			body := bluge.NewTextField("body", string(data)).WithAnalyzer(ja.Analyzer())
			doc := bluge.NewDocument(p).AddField(body)
			err = s.writer.Update(doc.ID(), doc)

			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s Service) Query(q string, top int) ([]QueryResult, error) {
	result := make([]QueryResult, 0)
	query := bluge.NewMatchQuery(q).SetAnalyzer(ja.Analyzer()).SetField("body")
	req := bluge.NewTopNSearch(top, query).WithStandardAggregations()
	reader, err := bluge.OpenReader(s.config)
	if err != nil {
		return result, err
	}
	iter, err := reader.Search(context.Background(), req)

	for {
		match, err := iter.Next()
		if err != nil {
			log.Fatalf("error iterator document matches: %v", err)
		}
		if match == nil {
			break
		}
		if err := match.VisitStoredFields(func(field string, value []byte) bool {
			result = append(result, QueryResult{filepath: string(value)})
			fmt.Printf("%s: %q\n", field, string(value))
			return true
		}); err != nil {
			log.Fatalf("error loading stored fields: %v", err)
		}
	}
	if err := reader.Close(); err != nil {
		log.Fatalf("error closing reader: %v", err)
		return result, err
	}
	return result, err
}
