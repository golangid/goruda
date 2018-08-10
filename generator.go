package goruda

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strings"
	"time"

	"text/template"

	"github.com/Masterminds/sprig"
	"gopkg.in/yaml.v2"
)

const (
	gorudaPacakages = "github.com/golangid/goruda"
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

func getDataType(origin string) string {

	switch origin {
	case "integer":
		return reflect.Int64.String()
	case "number":
		return reflect.Float64.String()
	}
	return origin
}

func extractMapToStrcutProperties(mapFieldStruct map[string]string) []Attribute {
	res := []Attribute{}
	for fieldName, fieldType := range mapFieldStruct {
		att := Attribute{
			Name: fieldName,
			Type: getDataType(fieldType),
		}
		res = append(res, att)
	}
	return res
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
		TimeStamp:  time.Now(),
		StructName: structName,
		Attributes: extractMapToStrcutProperties(mapFieldWithType),
	}
	filePath := fmt.Sprintf("%s/src/%s/template/struct_template.tpl", build.Default.GOPATH, gorudaPacakages)
	nameFile := path.Base(filePath)
	tmpl, err := template.New(nameFile).Funcs(sprig.TxtFuncMap()).ParseFiles(filePath)
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
