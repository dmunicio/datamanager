package models

type AssetBase struct {
	Type string `json:"type" doc:"Asset type" enum:"type1,type2"`
}

type AssetBaseResponse struct {
	Id string `json:"id" doc:"asset unique identifier"`
}

type AssetType1 struct {
	AssetBase
	Type string `json:"type" doc:"Asset type" enum:"type1"`
	Name string `json:"name" doc:"The name of the id"`
}

type AssetType2 struct {
	AssetBase
	Type   string  `json:"type" doc:"Asset type" enum:"type2"`
	Amount float64 `json:"amount"`
}

type AssetType1Response struct {
	AssetBaseResponse
	AssetType1
}

type AssetType2Response struct {
	AssetBaseResponse
	AssetType2
}
