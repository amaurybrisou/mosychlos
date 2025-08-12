package openai

import (
	"reflect"
	"slices"
	"strings"
	"time"
)

func BuildSchema[T any]() map[string]any {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	props := map[string]any{}
	req := []string{}

	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			jsonTag := f.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				continue
			}

			// Parse JSON tag
			tagParts := strings.Split(jsonTag, ",")
			fieldName := tagParts[0]

			// Skip empty field names
			if fieldName == "" {
				continue
			}

			// For OpenAI Responses API, fields with omitempty should not be included in schema at all
			// since ALL fields in properties must be in required array
			fieldType := f.Type
			isPointer := fieldType.Kind() == reflect.Pointer
			if isPointer {
				fieldType = fieldType.Elem()
			}

			hasOmitEmpty := slices.Contains(tagParts[1:], "omitempty")

			// Skip optional fields entirely for Responses API
			if hasOmitEmpty || isPointer {
				continue
			}

			schema := buildTypeSchema(f.Type)
			props[fieldName] = schema

			// For OpenAI Responses API, ALL properties must be required
			if fieldType.Kind() != reflect.Map {
				req = append(req, fieldName)
			}
		}
	}

	return map[string]any{
		"type":                 "object",
		"properties":           props,
		"required":             req,
		"additionalProperties": false,
	}

}

func buildTypeSchema(t reflect.Type) map[string]any {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}
	case reflect.Slice, reflect.Array:
		// For arrays, we need to specify the items type
		elemType := t.Elem()
		return map[string]any{
			"type":  "array",
			"items": buildTypeSchema(elemType),
		}
	case reflect.Struct:
		// Handle time.Time specially - it should be a string in JSON
		if t == reflect.TypeOf(time.Time{}) {
			return map[string]any{
				"type":   "string",
				"format": "date-time",
			}
		}

		// For structs, build the schema exactly like BuildSchema does for consistency
		props := map[string]any{}
		req := []string{}

		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			jsonTag := f.Tag.Get("json")
			if jsonTag == "" || jsonTag == "-" {
				continue
			}

			// Parse JSON tag exactly as in BuildSchema
			tagParts := strings.Split(jsonTag, ",")
			fieldName := tagParts[0]

			// Skip empty field names exactly as in BuildSchema
			if fieldName == "" {
				continue
			}

			// For OpenAI Responses API, fields with omitempty should not be included in schema at all
			fieldType := f.Type
			isPointer := fieldType.Kind() == reflect.Pointer
			if isPointer {
				fieldType = fieldType.Elem()
			}

			hasOmitEmpty := slices.Contains(tagParts[1:], "omitempty")

			// Skip optional fields entirely for Responses API
			if hasOmitEmpty || isPointer {
				continue
			}

			props[fieldName] = buildTypeSchema(f.Type)

			// For OpenAI Responses API, ALL properties must be required
			if fieldType.Kind() != reflect.Map {
				req = append(req, fieldName)
			}
		}

		return map[string]any{
			"type":                 "object",
			"properties":           props,
			"required":             req,
			"additionalProperties": false,
		}
	case reflect.Map:
		// For maps, we need to handle the additional properties schema
		valueType := t.Elem()
		return map[string]any{
			"type":                 "object",
			"additionalProperties": buildTypeSchema(valueType),
		}
	default:
		return map[string]any{"type": "string"}
	}
}
