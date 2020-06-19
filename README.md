# Goruda

Goruda is Golang CLI App to convert OpenAPI 3 Specifications
into simple boilerplate code. 

This app is created because the process of creating boilerplate 
HTTP server in Go is repetitive, so in order to reduce that kind of work, 
Goruda will read all your OpenAPI 3 specifications and convert it
into simple running HTTP Server.

## Requirements

- Go at least ver. 1.11
- Working OpenAPI 3 File

## How to Run

```bash
~ make build
~ ./goruda generate [path_to_openapi_file]
```

