package goruda_test

import (
	"testing"

	"github.com/golangid/goruda"
)

func TestGenerateStruct(t *testing.T) {
	gen := goruda.Goruda{
		PackageName:     "domain",
		TargetDirectory: "generated",
	}
	err := gen.Generate("./docs/menekel.yaml")
	if err != nil {
		t.Error(err)
	}
}
