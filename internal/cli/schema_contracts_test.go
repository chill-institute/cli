package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

type protoField struct {
	Name     string
	Type     string
	Repeated bool
	Optional bool
}

func TestLocalSchemaMatchesContractsProto(t *testing.T) {
	t.Parallel()

	path := strings.TrimSpace(os.Getenv("CHILLY_CONTRACTS_PROTO"))
	if path == "" {
		path = firstExistingPath(
			"../chill-contracts/proto/chill/v4/api.proto",
			"../../../chill-contracts/proto/chill/v4/api.proto",
		)
	}

	payload, err := os.ReadFile(path)
	if err != nil {
		if strings.TrimSpace(os.Getenv("CHILLY_REQUIRE_CONTRACTS")) == "1" {
			t.Fatalf("read contracts proto %q: %v", path, err)
		}
		t.Skipf("contracts proto not available at %q", path)
	}

	messages := parseProtoMessages(string(payload))
	assertTypeMatchesProtoMessage(t, typeReleaseInfo, messages["ReleaseInfo"])
	assertTypeMatchesProtoMessage(t, typeSearchResult, messages["SearchResult"])
	assertTypeMatchesProtoMessage(t, typeSearchResponse, messages["SearchResponse"])
}

func firstExistingPath(paths ...string) string {
	for _, path := range paths {
		cleaned := filepath.Clean(path)
		if _, err := os.Stat(cleaned); err == nil {
			return cleaned
		}
	}
	if len(paths) == 0 {
		return ""
	}
	return filepath.Clean(paths[0])
}

func assertTypeMatchesProtoMessage(t *testing.T, typeID string, protoFields []protoField) {
	t.Helper()

	schema, ok := typeSchemaRegistry[typeID]
	if !ok {
		t.Fatalf("missing local schema type %q", typeID)
	}
	if len(protoFields) == 0 {
		t.Fatalf("missing proto message for %q", typeID)
	}
	if len(schema.Fields) != len(protoFields) {
		t.Fatalf("%s field count = %d, want %d", typeID, len(schema.Fields), len(protoFields))
	}

	byName := make(map[string]schemaField, len(schema.Fields))
	for _, field := range schema.Fields {
		byName[field.Name] = field
	}

	for _, protoField := range protoFields {
		field, ok := byName[protoField.Name]
		if !ok {
			t.Fatalf("%s missing field %q", typeID, protoField.Name)
		}
		if field.Type != protoField.Type {
			t.Fatalf("%s.%s type = %q, want %q", typeID, field.Name, field.Type, protoField.Type)
		}
		if field.Repeated != protoField.Repeated {
			t.Fatalf("%s.%s repeated = %v, want %v", typeID, field.Name, field.Repeated, protoField.Repeated)
		}
		if field.Optional != protoField.Optional {
			t.Fatalf("%s.%s optional = %v, want %v", typeID, field.Name, field.Optional, protoField.Optional)
		}
		if field.JSONName != snakeToLowerCamel(field.Name) {
			t.Fatalf("%s.%s json_name = %q, want %q", typeID, field.Name, field.JSONName, snakeToLowerCamel(field.Name))
		}
	}
}

func parseProtoMessages(payload string) map[string][]protoField {
	messages := map[string][]protoField{}
	messagePattern := regexp.MustCompile(`(?s)message\s+(\w+)\s*\{(.*?)\n\}`)
	fieldPattern := regexp.MustCompile(`(?m)^\s*(optional\s+|repeated\s+)?(\w+)\s+(\w+)\s*=\s*\d+\s*;`)

	for _, messageMatch := range messagePattern.FindAllStringSubmatch(payload, -1) {
		name := messageMatch[1]
		body := messageMatch[2]
		fields := make([]protoField, 0)
		for _, fieldMatch := range fieldPattern.FindAllStringSubmatch(body, -1) {
			modifier := strings.TrimSpace(fieldMatch[1])
			fields = append(fields, protoField{
				Name:     fieldMatch[3],
				Type:     schemaTypeForProtoField(fieldMatch[2]),
				Repeated: modifier == "repeated",
				Optional: modifier == "optional",
			})
		}
		messages[name] = fields
	}
	return messages
}

func schemaTypeForProtoField(protoType string) string {
	switch protoType {
	case "string":
		return "string"
	case "int32", "int64":
		return "integer"
	case "bool":
		return "boolean"
	case "ReleaseInfo":
		return typeReleaseInfo
	case "SearchResult":
		return typeSearchResult
	default:
		return "chill.v4." + protoType
	}
}
