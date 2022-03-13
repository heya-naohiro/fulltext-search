package main

import (
	"context"
	"fmt"
	"log"

	"github.com/blugelabs/bluge"
	segment "github.com/blugelabs/bluge_segment_api"
	"github.com/ikawaha/blugeplugin/analysis/lang/ja"
)

func main() {
	config := bluge.InMemoryOnlyConfig()

	w, err := bluge.OpenWriter(config)
	if err != nil {
		log.Fatalf("error openning writer: %v", err)
	}

	defer w.Close()

	docs := NewDocuments()

	for _, doc := range docs {
		doc := doc
		if err := w.Update(doc.ID(), doc); err != nil {
			log.Fatalf("error updating document: %v", err)
		}

		fmt.Printf("indexed document with id: %s \n", doc.ID())
		doc.EachField(func(field segment.Field) {
			fmt.Printf("\t%s: %s\n", field.Name(), field.Value())
		})
	}

	r, err := w.Reader()
	if err != nil {
		log.Fatalf("error getting index reader: %v", err)
	}
	defer r.Close()

	q := "踊る人形"
	query := bluge.NewMatchQuery(q).SetAnalyzer(ja.Analyzer()).SetField("body")
	req := bluge.NewTopNSearch(10, query).WithStandardAggregations()
	fmt.Printf("query: search field %q, value %q\n", query.Field(), q)

	// search
	ite, err := r.Search(context.Background(), req)
	if err != nil {
		log.Fatalf("error executing search: %v", err)
	}

	for {
		match, err := ite.Next()
		if err != nil {
			log.Fatalf("error iterator document matches: %v", err)
		}
		if match == nil {
			break
		}
		if err := match.VisitStoredFields(func(field string, value []byte) bool {
			fmt.Printf("%s: %q\n", field, string(value))
			return true
		}); err != nil {
			log.Fatalf("error loading stored fields: %v", err)
		}
	}

}

func NewDocuments() []*bluge.Document {
	docs := []struct {
		ID     string
		Author string
		Text   string
	}{
		{
			ID:     "1:赤い蝋燭と人魚",
			Author: "小川未明",
			Text:   "人魚は南の方の海にばかり棲んでいるのではありません",
		},
		{
			ID:     "2:吾輩は猫である",
			Author: "夏目漱石",
			Text:   "吾輩は猫である。名前はまだない",
		},
		{
			ID:     "3:狐と踊れ",
			Author: "神林長平",
			Text:   "踊っているのでなければ踊らされているのだろうさ",
		},
		{
			ID:     "4:ダンスダンスダンス",
			Author: "村上春樹",
			Text:   "音楽の鳴っている間はとにかく踊り続けるんだ。おいらの言っていることはわかるかい？",
		},
	}

	var ret []*bluge.Document
	for _, v := range docs {
		auth := bluge.NewTextField("author", v.Author).WithAnalyzer(ja.Analyzer())
		body := bluge.NewTextField("body", v.Text).WithAnalyzer(ja.Analyzer())
		doc := bluge.NewDocument(v.ID).AddField(auth).AddField(body)
		ret = append(ret, doc)
	}
	return ret
}
