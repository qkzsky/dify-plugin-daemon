package parser

import (
	"encoding/json"

	"github.com/langgenius/dify-plugin-daemon/internal/types/validators"
)

func UnmarshalJson[T any](text string) (T, error) {
	return UnmarshalJsonBytes[T]([]byte(text))
}

func UnmarshalJsonBytes[T any](data []byte) (T, error) {
	var result T
	err := json.Unmarshal(data, &result)
	if err != nil {
		return result, err
	}

	if err := validators.GlobalEntitiesValidator.Struct(&result); err != nil {
		return result, err
	}

	return result, err
}

func MarshalJson[T any](data T) string {
	b, _ := json.Marshal(data)
	return string(b)
}

func MarshalJsonBytes[T any](data T) []byte {
	b, _ := json.Marshal(data)
	return b
}

func UnmarshalJsonBytes2Map(data []byte) (map[string]any, error) {
	result := map[string]any{}
	err := json.Unmarshal(data, &result)
	return result, err
}

func UnmarshalJson2Map(json string) (map[string]any, error) {
	return UnmarshalJsonBytes2Map([]byte(json))
}
