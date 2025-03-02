package main

import (
	"context"
	"datamanager/models"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"text/template"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"github.com/google/uuid"
)

type AssetPostRequest struct {
	Body interface{} `json:"body" oneOf:"AssetDocument,AssetPayment"`
}

type AssetPostResponse struct {
	Body struct {
		Message string `json:"message" doc:"Operation response"`
	}
}

type AssetGetRequest struct {
	Id string `path:"id" doc:"Asset id to retrieve"`
}

type AssetGetResponse struct {
	Body interface{} `json:"body" oneOf:"DocumentAssetResponse,PaymentAssetResponse"`
}

// API handler to process the request dynamically
func handleAssetPost(ctx context.Context, input *AssetPostRequest) (*AssetPostResponse, error) {
	// Ensure input.Body is actually a map
	bodyMap, ok := input.Body.(map[string]interface{})
	if !ok {
		return nil, huma.Error400BadRequest("Invalid JSON format")
	}

	// Extract the "type" field from the request body
	typeVal, exists := bodyMap["type"].(string)
	if !exists {
		return nil, huma.Error400BadRequest("Missing 'type' field in request")
	}
	id := uuid.New().String()
	jsonFile := id + ".json"
	bodyMap["id"] = id
	switch typeVal {
	case "type1":
		var doc models.AssetType1
		jsonData, _ := json.MarshalIndent(bodyMap, "", "  ") // Convert map to JSON
		if err := json.Unmarshal(jsonData, &doc); err != nil {
			return nil, huma.Error400BadRequest("Invalid Type format")
		}
		err := os.WriteFile(jsonFile, jsonData, 0644)
		if err != nil {
			return nil, huma.Error500InternalServerError("Error writing file:", err)
		}
	case "type2":
		var payment models.AssetType2
		jsonData, _ := json.Marshal(bodyMap) // Convert map to JSON
		if err := json.Unmarshal(jsonData, &payment); err != nil {
			return nil, huma.Error400BadRequest("Invalid Type format")
		}
		err := os.WriteFile(jsonFile, jsonData, 0644)
		if err != nil {
			return nil, huma.Error500InternalServerError("Error writing file:", err)
		}
		fmt.Println("DEBUG: Successfully parsed PaymentAsset")
	default:
		fmt.Println("DEBUG: Unknown type received:", typeVal)
		return nil, huma.Error400BadRequest("Unknown type: " + typeVal)
	}
	resp := &AssetPostResponse{}
	resp.Body.Message = "asset saved successfully to " + jsonFile
	return resp, nil

}

func handleAssetGet(ctx context.Context, input *AssetGetRequest) (*AssetGetResponse, error) {
	resp := &AssetGetResponse{}
	jsonFile := input.Id + ".json"
	file, err := os.Open(jsonFile)
	fmt.Println("opening file:", jsonFile)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil, huma.Error404NotFound("not found file " + jsonFile)
	}
	defer file.Close() // Ensure the file is closed when done
	jsonData, _ := os.ReadFile(jsonFile)
	var baseAsset models.AssetBase
	err = json.Unmarshal(jsonData, &baseAsset)
	if err != nil {
		return nil, huma.Error500InternalServerError("Error parsing file", err)
	}

	var asset interface{}
	switch baseAsset.Type {
	case "type1":
		var docAsset models.AssetType1Response
		json.Unmarshal(jsonData, &docAsset)
		asset = docAsset

	case "type2":
		var paymentAsset models.AssetType2Response
		json.Unmarshal(jsonData, &paymentAsset)
		asset = paymentAsset
	default:
		return nil, huma.Error500InternalServerError("Unknown asset type", err)

	}
	resp.Body = asset
	return resp, nil
}

func registerCustomDocsHandler(mux *http.ServeMux) {
	// Register custom docs handler
	mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		templatePath := filepath.Join("docs", "docs-template.html")
		tmpl, err := template.ParseFiles(templatePath)
		if err != nil {
			http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		err = tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func registerRoutes(api huma.API) {
	registry := api.OpenAPI().Components.Schemas
	type1Schema := registry.Schema(reflect.TypeOf(models.AssetType1{}), true, "")
	type2Schema := registry.Schema(reflect.TypeOf(models.AssetType2{}), true, "")
	requestSchema := &huma.Schema{
		OneOf: []*huma.Schema{
			type1Schema,
			type2Schema,
		},
	}
	requestSchema.Discriminator = &huma.Discriminator{PropertyName: "type"}
	requestBody := huma.RequestBody{
		Description: "Expected body",
		Content: map[string]*huma.MediaType{
			"application/json": {
				Schema: requestSchema,
			},
		},
	}

	huma.Register(api, huma.Operation{
		Method:      http.MethodPost,
		Path:        "/asset",
		Description: "Create an entry",
		RequestBody: &requestBody,
	}, handleAssetPost)

	type1ResponseSchema := registry.Schema(reflect.TypeOf(models.AssetType1Response{}), true, "")
	type2ResponseSchema := registry.Schema(reflect.TypeOf(models.AssetType2Response{}), true, "")
	assetGetResponseSchema := &huma.Schema{
		OneOf: []*huma.Schema{
			type1ResponseSchema,
			type2ResponseSchema,
		},
	}

	huma.Register(api, huma.Operation{
		Method:      http.MethodGet,
		Path:        "/asset/{id}",
		Description: "Retrieve an entry",
		Responses: map[string]*huma.Response{
			"200": {
				Content: map[string]*huma.MediaType{
					"application/json": {
						Schema: assetGetResponseSchema,
					},
				},
			},
			"400": {Description: "Invalid Request"},
		},
	}, handleAssetGet)
}

func main() {
	// Create a custom configuration for Huma that disables the built-in docs UI
	config := huma.DefaultConfig("Data Manager API", "1.0.0")
	config.DocsPath = ""
	// Create a new ServeMux
	mux := http.NewServeMux()
	// Register your custom docs handler BEFORE initializing the Huma API
	registerCustomDocsHandler(mux)
	// Create a Huma v2 API on top of the router
	api := humago.New(mux, config)

	registerRoutes(api)

	// Start the HTTP server to serve the API
	port := 8080
	portStr := strconv.Itoa(port)
	host := "http://localhost"
	docsEndpoint := "/docs"
	fmt.Printf("Starting server on :%s...\n", portStr)
	fmt.Printf("OpenAPI URL: %s:%s%s\n", host, portStr, docsEndpoint)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
