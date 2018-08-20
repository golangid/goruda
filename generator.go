package goruda

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"regexp"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/ghodss/yaml"
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

	components, ok := res["components"]
	if !ok {
		return ErrWrongYAMLFormat
	}
	err = processComponents(components)
	if err != nil {
		return err
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

func processPath(path interface{}) {

}

func isString(k interface{}) bool {
	switch k.(type) {
	case string:
		return true
	}
	return false
}

func processComponents(components interface{}) error {

	componentsMap := components.(map[string]interface{})

	schemasMap := componentsMap["schemas"].(map[string]interface{})
	return produceStruct(schemasMap)

}

func produceStruct(schemas map[string]interface{}) error {
	domainDatas := []DomainData{}

	for key, val := range schemas {
		domainData := DomainData{
			StructName: key,
		}
		valMap := val.(map[string]interface{})
		err := fillDomainData(&domainData, valMap)
		if err != nil {
			return err
		}

		domainDatas = append(domainDatas, domainData)

	}
	for _, data := range domainDatas {
		err := writeToFile(data)
		if err != nil {
			return err
		}
	}
	return nil
}

func fillDomainData(domainData *DomainData, attributes map[string]interface{}) error {

	properties := attributes["properties"]
	arrType := attributes["type"]
	if arrType != nil {
		err := resolveArrayType(domainData, attributes)
		if err != nil {
			return err
		}
	}
	propertiesMap := map[string]interface{}{}
	if properties != nil {
		propertiesMap = properties.(map[string]interface{})
	}
	arrAttribute := []Attribute{}
	for key, val := range propertiesMap {
		att := Attribute{
			Name: key,
		}
		valMap := val.(map[string]interface{})
		err := fillDataType(&att, valMap)
		if err != nil {
			return err
		}

		arrAttribute = append(arrAttribute, att)
	}
	domainData.Attributes = append(domainData.Attributes, arrAttribute...)
	return nil
}

func fillDataType(att *Attribute, dataType map[string]interface{}) error {
	typeAtt := dataType["type"]
	typeFormat := dataType["format"]
	if typeAtt == nil {
		return ErrWrongYAMLFormat
	}

	res, err := getDataType(typeAtt, typeFormat)
	if err != nil {
		return err
	}
	att.Type = res
	return nil
}
func getDataType(typeAtt, format interface{}) (string, error) {

	typeRes, ok := typeAtt.(string)
	if !ok {
		return "", ErrWrongYAMLFormat
	}

	switch typeRes {
	case "string":
		return reflect.String.String(), nil
	case "integer":
		if format != nil {
			return format.(string), nil
		}
		return reflect.Int64.String(), nil
	}
	return "", nil
}

func writeToFile(data DomainData) error {

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

	file, err := os.Create("generated/" + data.StructName + ".go")
	if err != nil {
		return err
	}

	defer file.Close()

	err = tmpl.Execute(file, data)
	if err != nil {
		return err
	}

	return nil
}

func resolveArrayType(domain *DomainData, attributes map[string]interface{}) error {

	typeVal, ok := attributes["type"].(string)
	if !ok {
		return ErrWrongYAMLFormat
	}

	if typeVal == "array" {
		re := regexp.MustCompile(`^#\/components\/schemas\/(.+)`)
		itemsMap := attributes["items"].(map[string]interface{})
		refVal := itemsMap["$ref"].(string)
		match := re.FindStringSubmatch(refVal)
		if len(match) < 2 {
			return ErrWrongYAMLFormat
		}
		att := Attribute{
			Name: fmt.Sprintf("%ss", match[1]),
			Type: fmt.Sprintf("[]%s", match[1]),
		}
		domain.Attributes = append(domain.Attributes, att)
	}

	return nil
}
