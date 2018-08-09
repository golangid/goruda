package goruda

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

func ReadJSONFile(jsonPath string) (map[string]interface{}, error) {
	jsonFile, err := os.OpenFile(jsonPath, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	defer jsonFile.Close()

	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{}, 0)
	err = json.Unmarshal(byteValue, &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}
