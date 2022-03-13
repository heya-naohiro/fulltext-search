package main

import (
	"fmt"
	"os"
	"testing"
)

func TestNewSerice(t *testing.T) {
	if _, err := os.Stat("./testdata/.blugeindex"); err == nil {
		if err := os.RemoveAll("./testdata/.blugeindex"); err != nil {
			t.Errorf("Remove Error %s", err)
		}
	}
	service, err := NewService("./testdata")
	if err != nil {
		t.Errorf("Create Service Error")
	}

	if err := service.CreateIndex(); err != nil {
		t.Errorf("Create Index Error")
	}

	if result, err := service.Query("名前", 2); err != nil || len(result) == 0 {
		t.Errorf("Query Error result: %d", len(result))

		if result[0].String() != "testdata/メモ1.txt" {
			t.Errorf("Query Result Error")
		}
		fmt.Println(result[0])
	}

}
