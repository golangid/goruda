package goruda

import (
	"errors"
	"testing"
)

func TestReadJSONFile(t *testing.T) {
	res, err := ReadJSONFile("json_file_test.json")
	if err != nil {
		t.Errorf("Test fail: %v", err)
	}

	if len(res) < 3 {
		t.Errorf("Test fail: %v", errors.New("Cannot read multiple field json"))
	}

	expectedResult := map[string]interface{}{
		"key": "val",
		"nested_key": map[string]interface{}{
			"in_nested_key": "nested_val",
		},
		"nested_key_with_array": []map[string]interface{}{
			{
				"array_key_1": "array_val_1",
			},
			{
				"array_key_2": "array_val_2",
			},
		},
	}

	if res["key"] != expectedResult["key"] {
		t.Errorf("Test fail: %v | wrong value: %v", errors.New("Result is not the same with expected value"), res["key"])
	}

	if res["nested_key"].(map[string]interface{})["in_nested_key"] != expectedResult["nested_key"].(map[string]interface{})["in_nested_key"] {
		t.Errorf("Test fail: %v | wrong value: %v", errors.New("Result is not the same with expected value"), res["in_nested_key"])
	}

	expectedResultMap := expectedResult["nested_key_with_array"].([]map[string]interface{})
	resMap := res["nested_key_with_array"].([]interface{})

	if len(expectedResultMap) != len(resMap) {
		t.Errorf("Test fail: %v ", errors.New("nested_with_array doesn't has same value"))
	}
}

func TestReadJSONFileNotFoundError(t *testing.T) {
	res, err := ReadJSONFile("json_file_tests.json")
	if err == nil {
		t.Errorf("Test fail: %v", errors.New("File is found"))
	}

	if res != nil {
		t.Errorf("Test fail: %v", errors.New("Result is not nil"))
	}
}

func TestReadJSONFileErrorFormat(t *testing.T) {
	res, err := ReadJSONFile("json_file_test_error.json")
	if err == nil {
		t.Errorf("Test fail: %v", errors.New("File is not error"))
	}

	if res != nil {
		t.Errorf("Test fail: %v", errors.New("Result is not nil"))
	}
}
