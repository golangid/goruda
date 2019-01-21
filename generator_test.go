package goruda_test

import (
	"testing"

	"github.com/golangid/goruda"
)

func TestGenerateStruct(t *testing.T) {
	err := goruda.Generate("./menekel.yaml")
	if err != nil {
		t.Error(err)
	}
}
