package goruda

import (
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/gobuffalo/packr"
	"os"
	"strings"
	"text/template"
	"time"

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
		case "object":
			err := generateStruct(k, v)
			if err != nil {
				return err
			}
		default:
			if len(v.Value.Properties) > 0 {
				err := generateStruct(k, v)
				if err != nil {
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
	box := packr.NewBox("./templates")
	str, err := box.FindString("struct_template.tpl")
	if err != nil {
		return err
	}
	tmpl, err := template.New("struct_template").Funcs(sprig.TxtFuncMap()).Parse(str)
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
