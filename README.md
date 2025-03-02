## About The Project

This is an example of a REST API in Go using [huma library](https://huma.rocks/) which generates automatically OpenAPI 3.1 definition and JSON Schemas.

The code is based on [this reddit post](https://www.reddit.com/r/golang/comments/1ajnhfb/go_122_builtin_routing_with_openapi_via_huma/)

## The challenge

This is a showcase for serving or receiving different schemas in the same endpoint by leveraging `oneOf` property.
It uses the `http.ServeMux` so no external routing package like Chi/Echo/Gin/Fiber is required.
It overrides the default /docs

## Getting Started

Install go 1.22 or later

Then just run:

```
go run main.go
```

Then go to [http://localhost:8080/docs](http://localhost:8080/docs)

You can access to the generated openapi.json: [http://localhost:8080/openapi.json](http://localhost:8080/openapi.json)
