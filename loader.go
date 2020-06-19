package goruda

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

func LoadSwaggerFile(swaggerFile string) *openapi3.Swagger {
	var err error
	var data []byte

	if strings.HasPrefix(swaggerFile, "delivery") {
		resp, err := http.Get(swaggerFile)
		if err != nil {
			log.Fatal(err)
		}

		data, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		data, err = ioutil.ReadFile(swaggerFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	loader := openapi3.NewSwaggerLoader()
	var swagger *openapi3.Swagger
	if strings.HasSuffix(swaggerFile, ".yaml") || strings.HasSuffix(swaggerFile, ".yml") {
		swagger, err = loader.LoadSwaggerFromYAMLData(data)
	} else {
		swagger, err = loader.LoadSwaggerFromData(data)
	}
	if err != nil {
		log.Fatal(err)
	}
	return swagger
}
