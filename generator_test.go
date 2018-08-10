package goruda

import "testing"

func TestGenerateStructFromYAML(t *testing.T) {
	err := GenerateStructFromYAML("yaml_struct_test.yaml")
	if err != nil {
		t.Error(err)
	}
}
