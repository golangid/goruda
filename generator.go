package goruda

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig"
	"github.com/gobuffalo/packr"

	"github.com/getkin/kin-openapi/openapi3"
)

const (
	gorudaPacakages = "github.com/golangid/goruda"
)

type Goruda struct {
	PackageName     string
	TargetDirectory string
}

func (g Goruda) Generate(swaggerFile string) error {
	swagger := LoadSwaggerFile(swaggerFile)
	if err := g.generateStructs(swagger); err != nil {
		return err
	}

	if err := g.generateCommandPackage(); err != nil {
		return err
	}
	return nil
}

func (g Goruda) generateStructs(swagger *openapi3.Swagger) error {
	for k, v := range swagger.Components.Schemas {
		t := v.Value.Type
		switch t {
		case "object":
			if err := g.generateStruct(k, v); err != nil {
				return err
			}
		case "array":
			if err := g.generateSliceStruct(k, v); err != nil {
				return err
			}
		default:
			if len(v.Value.Properties) > 0 {
				if err := g.generateStruct(k, v); err != nil {
					return err
				}
				continue
			}
		}

	}

	if err := g.generateServiceFile(g.retrieveAbstraction(swagger.Paths)); err != nil {
		return err
	}

	if err := g.generateHTTPDeliveryFile(HTTPData{
		TimeStamp:   time.Now(),
		PackageName: g.PackageName,
		ServiceName: "Service",
		Methods:     g.retrieveHTTPDeliveryData(swagger.Paths),
	}); err != nil {
		return err
	}
	return nil
}

// this function return schema type and bool status for polymorphism
func (g Goruda) getType(schema *openapi3.SchemaRef, schemaTitle ...string) (string, bool) {
	if schema.Ref != "" {
		return strings.Split(schema.Ref, "/")[3], false
	}
	if (len(schema.Value.OneOf) > 0 ||
		len(schema.Value.AnyOf) > 0 ||
		len(schema.Value.AllOf) > 0) && len(schemaTitle) > 0 {

		if len(schema.Value.OneOf) > 0 {
			return fmt.Sprintf("ChildOf%v", schemaTitle[0]), true
		}

		return "interface{}", false
	}

	switch schema.Value.Type {
	case "integer":
		if schema.Value.Format != "" {
			return schema.Value.Format, false
		}
		return "int", false
	case "string":
		if format := schema.Value.Format; format != "" {
			switch format {
			case "date-time", "date":
				return "time.Time", false
			}
		}
		return "string", false
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
		return "map[string]interface{}", false
	case "array":
		if items := schema.Value.Items; items != nil {
			schemaType, _ := g.getType(items)
			return fmt.Sprintf("[]%s", schemaType), false
		}
		// TODO: (by bxcodec)
		// Add Specific conditions for array of objects
	}
	return schema.Value.Type, false
}

func (g Goruda) generateSliceStruct(name string, schema *openapi3.SchemaRef) error {
	dmData := DomainData{
		StructName:  name,
		TimeStamp:   time.Now(),
		Packagename: g.PackageName,
	}

	imports := map[string]Import{}

	schemType, isPolymorph := g.getType(schema.Value.Items, name)
	if isPolymorph {
		if err := g.generatePolymorphStruct(schemType, schema.Value.Items); err != nil {
			return err
		}
	}
	setImports(schemType, imports)

	dmData.Imports = imports
	dmData.SliceData = SliceData{
		Type: schemType,
	}
	return g.generateFile(dmData)
}

func (g Goruda) generateStruct(name string, schema *openapi3.SchemaRef) error {
	dmData := DomainData{
		StructName:  name,
		TimeStamp:   time.Now(),
		Packagename: g.PackageName,
	}
	attributes := []Attribute{}
	imports := map[string]Import{}
	for k, v := range schema.Value.Properties {
		schemType, isPolymorph := g.getType(v, name)
		if isPolymorph {
			if err := g.generatePolymorphStruct(schemType, v); err != nil {
				return err
			}
		}
		att := Attribute{
			Name: k,
			Type: schemType,
		}
		setImports(schemType, imports)
		attributes = append(attributes, att)
	}
	dmData.Attributes = attributes
	dmData.Imports = imports
	return g.generateFile(dmData)
}

func (g Goruda) generatePolymorphStruct(name string, schema *openapi3.SchemaRef) error {
	attributes := []Attribute{}
	for _, ref := range schema.Value.OneOf {
		if ref.Ref == "" {
			firstLetter := ref.Value.Type[0]
			attributes = append(attributes, Attribute{
				Name: strings.ToUpper(string(firstLetter)) + ref.Value.Type[1:],
				Type: ref.Value.Type,
			})
			continue
		}
		attributes = append(attributes, Attribute{
			Name: "",
			Type: strings.Split(ref.Ref, "/")[3],
		})
	}

	dmData := DomainData{
		StructName:  name,
		TimeStamp:   time.Now(),
		Packagename: g.PackageName,
		IsPolymorph: true,
	}

	imports := map[string]Import{}
	dmData.Attributes = attributes
	dmData.Imports = imports
	return g.generateFile(dmData)
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

func (g Goruda) generateStructFile(data DomainData) error {
	box := packr.NewBox("./templates")
	str, err := box.FindString("struct_template.tpl")
	if err != nil {
		return err
	}
	tmpl, err := template.New("struct_template").Funcs(sprig.TxtFuncMap()).Parse(str)
	if err != nil {
		return err
	}

	if _, err := os.Stat(g.TargetDirectory); os.IsNotExist(err) {
		if err = os.Mkdir(g.TargetDirectory, os.ModePerm); err != nil {
			return err
		}
	}

	file, err := os.Create(g.TargetDirectory + "/" + data.StructName + ".go")
	if err != nil {
		return err
	}
	defer file.Close()

	var buf bytes.Buffer

	if err = tmpl.Execute(&buf, data); err != nil {
		return err
	}

	formattedString, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}

	_, err = file.WriteString(string(formattedString))
	if err != nil {
		return err
	}
	return nil
}

func (g Goruda) generateServiceFile(data AbstractionData) error {
	box := packr.NewBox("./templates")
	str, err := box.FindString("service_template.tpl")
	if err != nil {
		return err
	}
	tmpl, err := template.New("service_template").Funcs(sprig.TxtFuncMap()).Parse(str)
	if err != nil {
		return err
	}

	if _, err := os.Stat(g.TargetDirectory); os.IsNotExist(err) {
		if err = os.Mkdir(g.TargetDirectory, os.ModePerm); err != nil {
			return err
		}
	}

	file, err := os.Create(g.TargetDirectory + "/" + data.Name + ".go")
	if err != nil {
		return err
	}
	defer file.Close()
	if err = tmpl.Execute(file, data); err != nil {
		return err
	}

	str, err = box.FindString("service_test_template.tpl")
	if err != nil {
		return err
	}
	tmpl, err = template.New("service_test_template").Funcs(sprig.TxtFuncMap()).Parse(str)
	if err != nil {
		return err
	}

	file, err = os.Create("generated/" + data.Name + "_test.go")
	if err != nil {
		return err
	}
	defer file.Close()
	if err = tmpl.Execute(file, data); err != nil {
		return err
	}
	return nil
}

func (g Goruda) generateHTTPDeliveryFile(data HTTPData) error {
	box := packr.NewBox("./templates")
	str, err := box.FindString("delivery_http_template.tpl")
	if err != nil {
		return err
	}
	tmpl, err := template.New("delivery_http_template").Funcs(sprig.TxtFuncMap()).Parse(str)
	if err != nil {
		return err
	}

	if _, err := os.Stat("generated/internal"); os.IsNotExist(err) {
		if err = os.Mkdir("generated/internal", os.ModePerm); err != nil {
			return err
		}
	}

	if _, err := os.Stat("generated/internal/delivery"); os.IsNotExist(err) {
		if err = os.Mkdir("generated/internal/delivery", os.ModePerm); err != nil {
			return err
		}
	}

	file, err := os.Create("generated/internal/delivery/http.go")
	if err != nil {
		return err
	}
	defer file.Close()
	if err = tmpl.Execute(file, data); err != nil {
		return err
	}
	return nil
}

func (g Goruda) generateCommandPackage() error {
	box := packr.NewBox("./templates")
	str, err := box.FindString("command_template.tpl")
	if err != nil {
		return err
	}
	tmpl, err := template.New("command_template").Funcs(sprig.TxtFuncMap()).Parse(str)
	if err != nil {
		return err
	}

	if _, err := os.Stat(g.TargetDirectory + "/cmd"); os.IsNotExist(err) {
		if err = os.Mkdir(g.TargetDirectory + "/cmd", os.ModePerm); err != nil {
			return err
		}
	}

	if _, err := os.Stat(g.TargetDirectory + "/cmd/" + g.PackageName); os.IsNotExist(err) {
		if err = os.Mkdir(g.TargetDirectory + "/cmd/" + g.PackageName, os.ModePerm); err != nil {
			return err
		}
	}

	file, err := os.Create(g.TargetDirectory + "/cmd/command.go")
	if err != nil {
		return err
	}
	defer file.Close()
	if err = tmpl.Execute(file, g); err != nil {
		return err
	}

	str, err = box.FindString("main_function_template.tpl")
	if err != nil {
		return err
	}
	tmpl, err = template.New("main_function_template").Funcs(sprig.TxtFuncMap()).Parse(str)
	if err != nil {
		return err
	}

	file, err = os.Create(g.TargetDirectory + "/cmd/" + g.PackageName + "/main.go")
	if err != nil {
		return err
	}
	defer file.Close()
	if err = tmpl.Execute(file, g); err != nil {
		return err
	}
	return nil
}

func (g Goruda) generateFile(data DomainData) error {
	if err := g.generateStructFile(data); err != nil {
		return err
	}
	// TODO: (by bxcodec)
	// Another generation will place here
	return nil
}

func (g Goruda) retrieveAbstraction(paths openapi3.Paths) AbstractionData {
	methodsWithParam := map[string]ListOfAttributes{}
	for _, item := range paths {
		if item.Get != nil {
			name := item.Get.OperationID
			params := Attributes{}
			returnValues := Attributes{}
			for code, resp := range item.Get.Responses {
				if code == "200" {
					t, _ := g.getType(resp.Value.Content.Get("application/json").Schema)
					returnValues = append(returnValues, Attribute{
						Type: t,
					})
				}
			}
			for _, parameter := range item.Get.Parameters {
				schemaType, _ := g.getType(parameter.Value.Schema)
				isRequired := false
				switch parameter.Value.In {
				case "path":
					isRequired = true
				case "query":
				}
				params = append(params, Attribute{
					Name:       parameter.Value.Name,
					Type:       schemaType,
					IsRequired: isRequired,
				})
			}

			methodsWithParam[name] = ListOfAttributes{
				Attributes:  params,
				ReturnValue: returnValues,
			}
		}

		if item.Post != nil {
			name := item.Post.OperationID
			params := Attributes{}
			returnValues := Attributes{}
			for code, resp := range item.Post.Responses {
				if code == "201" {
					t, _ := g.getType(resp.Value.Content.Get("application/json").Schema)
					returnValues = append(returnValues, Attribute{
						Type: t,
					})
				}
			}

			schemaType, _ := g.getType(item.Post.RequestBody.Value.Content.
				Get("application/json").Schema)
			firstLetter := schemaType[0]
			params = append(params, Attribute{
				Name: strings.ToLower(string(firstLetter)) + schemaType[1:],
				Type: schemaType,
			})

			methodsWithParam[name] = ListOfAttributes{
				Attributes:  params,
				ReturnValue: returnValues,
			}
		}

		if item.Put != nil {
			name := item.Put.OperationID
			params := Attributes{}
			returnValues := Attributes{}
			for code, resp := range item.Put.Responses {
				if code == "200" {
					t, _ := g.getType(resp.Value.Content.Get("application/json").Schema)
					returnValues = append(returnValues, Attribute{
						Type: t,
					})
				}
			}

			for _, parameter := range item.Put.Parameters {
				schemaType, _ := g.getType(parameter.Value.Schema)
				isRequired := false
				switch parameter.Value.In {
				case "path":
					isRequired = true
				case "query":
				}
				params = append(params, Attribute{
					Name:       parameter.Value.Name,
					Type:       schemaType,
					IsRequired: isRequired,
				})
			}

			schemaType, _ := g.getType(item.Put.RequestBody.Value.Content.
				Get("application/json").Schema)
			firstLetter := schemaType[0]
			params = append(params, Attribute{
				Name: strings.ToLower(string(firstLetter)) + schemaType[1:],
				Type: schemaType,
			})

			methodsWithParam[name] = ListOfAttributes{
				Attributes:  params,
				ReturnValue: returnValues,
			}
		}

		if item.Delete != nil {
			name := item.Delete.OperationID
			params := Attributes{}
			returnValues := Attributes{}
			for code, resp := range item.Delete.Responses {
				switch code {
				case "404":
					t, _ := g.getType(resp.Value.Content.Get("application/json").Schema)
					returnValues = append(returnValues, Attribute{
						Type: t,
					})
				case "204":
				}
			}
			for _, parameter := range item.Delete.Parameters {
				schemaType, _ := g.getType(parameter.Value.Schema)
				params = append(params, Attribute{
					Name: parameter.Value.Name,
					Type: schemaType,
				})
			}

			methodsWithParam[name] = ListOfAttributes{
				Attributes:  params,
				ReturnValue: returnValues,
			}
		}
	}

	return AbstractionData{
		PackageName: g.PackageName,
		Name:        "Service",
		Methods:     methodsWithParam,
	}
}

func (g Goruda) retrieveHTTPDeliveryData(paths openapi3.Paths) map[string]HTTPMethods {
	httpMethods := map[string]HTTPMethods{}
	for key, item := range paths {
		if item.Get != nil {
			name := item.Get.OperationID
			params := Attributes{}
			returnValues := Attributes{}
			for code, resp := range item.Get.Responses {
				if code == "200" {
					t, _ := g.getType(resp.Value.Content.Get("application/json").Schema)
					returnValues = append(returnValues, Attribute{
						Type: t,
					})
				}
			}
			for _, parameter := range item.Get.Parameters {
				schemaType, _ := g.getType(parameter.Value.Schema)
				params = append(params, Attribute{
					Name: parameter.Value.Name,
					Type: schemaType,
				})
			}

			h := HTTPMethods{
				Path:        key,
				MethodsName: "GET",
				Data: ListOfAttributes{
					Attributes:  params,
					ReturnValue: returnValues,
				},
			}
			httpMethods[name] = h
		}

		if item.Post != nil {
			name := item.Post.OperationID
			params := Attributes{}
			returnValues := Attributes{}
			for code, resp := range item.Post.Responses {
				if code == "201" {
					t, _ := g.getType(resp.Value.Content.Get("application/json").Schema)
					returnValues = append(returnValues, Attribute{
						Type: t,
					})
				}
			}

			schemaType, _ := g.getType(item.Post.RequestBody.Value.Content.
				Get("application/json").Schema)
			firstLetter := schemaType[0]
			params = append(params, Attribute{
				Name: strings.ToLower(string(firstLetter)) + schemaType[1:],
				Type: schemaType,
			})

			httpMethods[name] = HTTPMethods{
				Path: key,
				MethodsName: "POST",
				Data: ListOfAttributes{
					Attributes:  params,
					ReturnValue: returnValues,
				},
			}
		}

		if item.Put != nil {
			name := item.Put.OperationID
			params := Attributes{}
			returnValues := Attributes{}
			for code, resp := range item.Put.Responses {
				if code == "200" {
					t, _ := g.getType(resp.Value.Content.Get("application/json").Schema)
					returnValues = append(returnValues, Attribute{
						Type: t,
					})
				}
			}

			for _, parameter := range item.Put.Parameters {
				schemaType, _ := g.getType(parameter.Value.Schema)
				isRequired := false
				switch parameter.Value.In {
				case "path":
					isRequired = true
				case "query":
				}
				params = append(params, Attribute{
					Name:       parameter.Value.Name,
					Type:       schemaType,
					IsRequired: isRequired,
				})
			}

			schemaType, _ := g.getType(item.Put.RequestBody.Value.Content.
				Get("application/json").Schema)
			firstLetter := schemaType[0]
			params = append(params, Attribute{
				Name: strings.ToLower(string(firstLetter)) + schemaType[1:],
				Type: schemaType,
			})

			httpMethods[name] = HTTPMethods{
				Path:        key,
				MethodsName: "PUT",
				Data:        ListOfAttributes{
					Attributes:  params,
					ReturnValue: returnValues,
				},
			}
		}

		if item.Delete != nil {
			name := item.Delete.OperationID
			params := Attributes{}
			returnValues := Attributes{}
			for code, resp := range item.Delete.Responses {
				switch code {
				case "404":
					t, _ := g.getType(resp.Value.Content.Get("application/json").Schema)
					returnValues = append(returnValues, Attribute{
						Type: t,
					})
				case "204":
				}
			}
			for _, parameter := range item.Delete.Parameters {
				schemaType, _ := g.getType(parameter.Value.Schema)
				params = append(params, Attribute{
					Name: parameter.Value.Name,
					Type: schemaType,
				})
			}

			httpMethods[name] = HTTPMethods{
				Path:        key,
				MethodsName: "DELETE",
				Data:        ListOfAttributes{
					Attributes:  params,
					ReturnValue: returnValues,
				},
			}
		}
	}
	return httpMethods
}
