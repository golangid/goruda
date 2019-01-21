package goruda

import (
	"fmt"
	"go/build"
	"os"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/getkin/kin-openapi/openapi3"
)

const (
	gorudaPacakages = "github.com/golangid/goruda"
)

func Generate(swaggerFile string) error {
	swagger := LoadSwaggerFile(swaggerFile)
	return generateStructs(swagger)
}

func generateStructs(swagger *openapi3.Swagger) error {

	for k, v := range swagger.Components.Schemas {
		t := v.Value.Type
		switch t {
		case "array":
			fmt.Println("Array In")
		case "object":
			err := generateStruct(k, v)
			if err != nil {
				return err
			}
			fmt.Println("========")
		default:
			if len(v.Value.Properties) > 0 {
				err := generateStruct(k, v)
				if err != nil {
					return err
				}
				fmt.Println("========")
				continue
			}
			fmt.Println("Default In, ", t)
		}

	}
	return nil
}

func getType(schema *openapi3.SchemaRef) string {

	if schema.Ref != "" {
		// TODO: (by bxcodec)
		// Add specific jobs to finish that using ref
	}
	switch schema.Value.Type {
	case "integer":
		if schema.Value.Format != "" {
			return schema.Value.Format
		}
		return "int"
	case "object":
		return "interface{}"
	case "array":
		if items := schema.Value.Items; items != nil {
			return fmt.Sprintf("[]%s", getType(items))
		}
		// TODO: (by bxcodec)
		// Add Specific conditions for array of objects
	}
	return schema.Value.Type
}

func generateStruct(name string, schema *openapi3.SchemaRef) error {
	dmData := DomainData{
		StructName:  name,
		TimeStamp:   time.Now(),
		Packagename: "domain",
	}
	attributes := []Attribute{}
	fmt.Println("NAME : ", name)
	for k, v := range schema.Value.Properties {
		att := Attribute{
			Name: strings.ToTitle(k),
			Type: getType(v),
		}
		attributes = append(attributes, att)
	}
	dmData.Attributes = attributes

	return generateFile(dmData)
}

func generateStructFile(data DomainData) error {
	filePath := fmt.Sprintf("%s/src/%s/templates/struct_template.tpl", build.Default.GOPATH, gorudaPacakages)
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

func generateFile(data DomainData) error {
	err := generateStructFile(data)
	if err != nil {
		return err
	}
	// TODO: (by bxcodec)
	// Another generation will place here
	return nil
}
