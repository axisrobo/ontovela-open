// Package contract guards SDK-to-OpenAPI parity. Every OpenAPI operationId
// must have a corresponding method in each SDK client.
package contract

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

type openAPISpec struct {
	Paths map[string]map[string]struct {
		OperationID string `yaml:"operationId"`
	} `yaml:"paths"`
}

func TestSDKMethodsCoverEveryOpenAPIOperation(t *testing.T) {
	root := filepath.Clean("../")
	specData, err := os.ReadFile(filepath.Join(root, "api", "openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	var spec openAPISpec
	if err := yaml.Unmarshal(specData, &spec); err != nil {
		t.Fatal(err)
	}
	operations := make([]string, 0)
	for _, methods := range spec.Paths {
		for _, method := range methods {
			if method.OperationID != "" {
				operations = append(operations, method.OperationID)
			}
		}
	}
	if len(operations) == 0 {
		t.Fatal("no operations found in OpenAPI")
	}

	assertGoMethods(t, root, operations)
	assertPythonMethods(t, root, operations)
	assertTypeScriptMethods(t, root, operations)
}

func assertGoMethods(t *testing.T, root string, operations []string) {
	t.Helper()
	source, err := os.ReadFile(filepath.Join(root, "sdk", "go", "client.go"))
	if err != nil {
		t.Fatal(err)
	}
	methods := methodSet(string(source), `func \(c \*Client\) (\w+)\(`)
	for _, operation := range operations {
		expected := upperFirst(operation)
		if !methods[expected] {
			t.Errorf("Go SDK missing method %q for operationId %q", expected, operation)
		}
	}
}

func assertPythonMethods(t *testing.T, root string, operations []string) {
	t.Helper()
	source, err := os.ReadFile(filepath.Join(root, "sdk", "python", "ontovela", "client.py"))
	if err != nil {
		t.Fatal(err)
	}
	methods := methodSet(string(source), `(?m)^    def (\w+)\(self`)
	for _, operation := range operations {
		expected := toSnake(operation)
		if !methods[expected] {
			t.Errorf("Python SDK missing method %q for operationId %q", expected, operation)
		}
	}
}

func assertTypeScriptMethods(t *testing.T, root string, operations []string) {
	t.Helper()
	source, err := os.ReadFile(filepath.Join(root, "sdk", "typescript", "src", "client.ts"))
	if err != nil {
		t.Fatal(err)
	}
	methods := methodSet(string(source), `  async (\w+)\(`)
	for _, operation := range operations {
		if !methods[operation] {
			t.Errorf("TypeScript SDK missing method %q for operationId %q", operation, operation)
		}
	}
}

func methodSet(source, pattern string) map[string]bool {
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(source, -1)
	set := make(map[string]bool, len(matches))
	for _, match := range matches {
		set[match[1]] = true
	}
	return set
}

func upperFirst(value string) string {
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

func toSnake(value string) string {
	re := regexp.MustCompile(`[A-Z]`)
	return strings.ToLower(re.ReplaceAllStringFunc(value, func(letter string) string {
		return "_" + strings.ToLower(letter)
	}))
}
