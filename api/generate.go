package generate

import "embed"

//go:generate go tool oapi-codegen -generate types -package api -o generated/types.gen.go spec/swagger.yaml

//go:embed swagger/* spec/swagger.yaml
var APIEmbeddedFiles embed.FS
