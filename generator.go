package goruda

import (
	"io/ioutil"
	"os"
	"strings"

	"text/template"

	"gopkg.in/yaml.v2"
)

// GenerateStructFromYAML is function to generate struct entity based on yaml file
func GenerateStructFromYAML(filePath string) error {
	byteValue, err := readFile(filePath)
	if err != nil {
		return err
	}

	res := make(map[string]interface{})
	err = yaml.Unmarshal(byteValue, &res)
	if err != nil {
		return err
	}

	itemDefinitions := res["definitions"]
	itemDefinitionsValue, ok := itemDefinitions.(map[interface{}]interface{})
	if !ok {
		return ErrWrongYAMLFormat
	}

	structName := make([]string, 0)

	for key := range itemDefinitionsValue {
		value, ok := key.(string)
		if !ok {
			return ErrWrongYAMLFormat
		}
		structName = append(structName, value)
	}

	for _, name := range structName {
		properties := itemDefinitionsValue[name]
		propertiesCasted, ok := properties.(map[interface{}]interface{})
		if !ok {
			return ErrWrongYAMLFormat
		}

		if propertiesCasted["type"] == "object" {
			err := produceStruct(name, propertiesCasted["properties"].(map[interface{}]interface{}))
			if err != nil {
				return err
			}

		}
	}

	return nil
}

func readFile(filePath string) ([]byte, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	defer file.Close()
	return ioutil.ReadAll(file)
}

func produceStruct(name string, properties map[interface{}]interface{}) error {
	fieldDescription := map[string]interface{}{}
	fieldWithType := map[string]string{}
	resultStructSpecification := map[string]interface{}{}
	for name, property := range properties {
		nameValue, ok := name.(string)
		if !ok {
			return ErrWrongYAMLFormat
		}
		fieldDescription[nameValue] = property
	}

	name = strings.Title(name)

	for fieldName, property := range fieldDescription {
		propertyCasted, ok := property.(map[interface{}]interface{})
		if !ok {
			return ErrWrongYAMLFormat
		}

		typeString, ok := propertyCasted["type"].(string)
		if !ok {
			return ErrWrongYAMLFormat
		}

		if typeString == "object" {
			err := produceStruct(fieldName, propertyCasted["properties"].(map[interface{}]interface{}))
			if err != nil {
				return err
			}
		}

		fieldName = strings.Title(fieldName)
		fieldWithType[fieldName] = typeString
		resultStructSpecification[name] = fieldWithType
	}

	return writeGeneratedStructToFile(resultStructSpecification)
}

func writeGeneratedStructToFile(structValue map[string]interface{}) error {
	structName := ""
	for key := range structValue {
		structName = key
	}

	mapFieldWithType, ok := structValue[structName].(map[string]string)
	if !ok {
		return ErrWrongYAMLFormat
	}

	data := DomainData{
		StructName:       structName,
		StructProperties: mapFieldWithType,
	}

	tmpl, err := template.New("test").Parse("package menekel \n\ntype {{.StructName}} struct " +
		"{ \n {{range $key, $value := .StructProperties}} {{$key}} {{if eq $value \"integer\"}} " +
		"int64 {{else if eq $value \"object\"}} {{$key}} {{else}} {{$value}} {{end}}\n {{end}} }")

	if err != nil {
		return err
	}

	if _, err := os.Stat("generated"); os.IsNotExist(err) {
		err = os.Mkdir("generated", os.ModePerm)
		if err != nil {
			return err
		}
	}

	file, err := os.Create("generated/" + structName + ".go")
	if err != nil {
		return err
	}

	defer file.Close()

	err = tmpl.Execute(file, data)
	if err != nil {
		return err
	}

	//TO DO
	//need to write to file

	return nil
}
