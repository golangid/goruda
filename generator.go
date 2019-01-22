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
		case "object":
			if err := generateStruct(k, v); err != nil {
				return err
			}
		default:
			if len(v.Value.Properties) > 0 {
				if err := generateStruct(k, v); err != nil {
					return err
				}
				continue
			}
		}

	}
	return nil
}

func getType(schema *openapi3.SchemaRef) string {
	if schema.Ref != "" {
		return strings.Split(schema.Ref, "/")[3]
	}
	if len(schema.Value.OneOf) > 0 ||
		len(schema.Value.AnyOf) > 0 ||
		len(schema.Value.AllOf) > 0 {
		// TODO: (by bxcodec)
		// It's hard to define if it comes to this kind of data:
		//  - oneOf
		//  - allOf
		//  - anyOf
		// Just return an plain interface{} let the developer decide later what should it best to this data types
		return "interface{}"
	}

	switch schema.Value.Type {
	case "integer":
		if schema.Value.Format != "" {
			return schema.Value.Format
		}
		return "int"
	case "string":
		if format := schema.Value.Format; format != "" {
			switch format {
			case "date-time", "date":
				return "time.Time"
			}
		}
		return "string"
	case "object":
		// TODO: (by bxcodec)
		// This section temporary I just send map[string]interface{}
		// Based on the condition that I believe, if it was an embedded object:
		// For example, this Article schema.
		// ```
		// Article:
		//  properties:
		// 	  publisher:
		// 	   	type: object
		// 	   	properties:
		// 		 id:
		// 		   type: string
		// 		 name:
		// 		   type: string
		// ```
		// That means, that publisher's object is not necesessary to have in the code as a struct.
		// If it important to have as a struct it must defined using `$ref` then.
		// But I'm a bit confused to use between interface{} or explicitly using map[string]interface{} will decide later
		// after a real test
		return "map[string]interface{}"
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
	imports := map[string]Import{}
	for k, v := range schema.Value.Properties {
		att := Attribute{
			Name: k,
			Type: getType(v),
		}
		setImports(getType(v), imports)
		attributes = append(attributes, att)
	}
	dmData.Attributes = attributes
	dmData.Imports = imports
	return generateFile(dmData)
}

func setImports(dataType string, imports map[string]Import) {
	switch dataType {
	case "time.Time":
		imports["time"] = Import{
			Alias: "time",
			Path:  "time",
		}
	}
}

func generateStructFile(data DomainData) error {
	filePath := fmt.Sprintf("%s/src/%s/templates/struct_template.tpl", build.Default.GOPATH, gorudaPacakages)
	nameFile := path.Base(filePath)
	tmpl, err := template.New(nameFile).Funcs(sprig.TxtFuncMap()).ParseFiles(filePath)
	if err != nil {
		return err
	}

	if _, err := os.Stat("generated"); os.IsNotExist(err) {
		if err = os.Mkdir("generated", os.ModePerm); err != nil {
			return err
		}
	}

	file, err := os.Create("generated/" + data.StructName + ".go")
	if err != nil {
		return err
	}
	defer file.Close()
	if err = tmpl.Execute(file, data); err != nil {
		return err
	}
	return nil
}

func generateFile(data DomainData) error {
	if err := generateStructFile(data); err != nil {
		return err
	}
	// TODO: (by bxcodec)
	// Another generation will place here
	return nil
}
