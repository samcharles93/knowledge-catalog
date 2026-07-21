package okf

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// OpenAPIExtractor parses an OpenAPI 3.0 / Swagger spec (JSON or YAML) into
// one API concept per operation.
type OpenAPIExtractor struct {
	SpecPath string
}

var _ Extractor = OpenAPIExtractor{}

var openAPIMethods = []string{"get", "post", "put", "delete", "patch"}

var openAPIOpIDCleanRE = regexp.MustCompile(`[^A-Za-z0-9_-]`)

func (e OpenAPIExtractor) ExtractConcepts() (map[string]Document, error) {
	data, err := os.ReadFile(e.SpecPath)
	if err != nil {
		return nil, err
	}

	var spec map[string]any
	if ext := strings.ToLower(filepath.Ext(e.SpecPath)); ext == ".yaml" || ext == ".yml" {
		err = yaml.Unmarshal(data, &spec)
	} else {
		err = json.Unmarshal(data, &spec)
	}
	if err != nil {
		return nil, err
	}

	info, _ := spec["info"].(map[string]any)
	version := stringFrontmatterOr(info["version"], "1.0.0")

	paths, _ := spec["paths"].(map[string]any)
	pathKeys := make([]string, 0, len(paths))
	for p := range paths {
		pathKeys = append(pathKeys, p)
	}
	sort.Strings(pathKeys)

	now := time.Now().UTC().Format(time.RFC3339)
	concepts := map[string]Document{}

	for _, pathStr := range pathKeys {
		pathItem, ok := paths[pathStr].(map[string]any)
		if !ok {
			continue
		}
		for _, method := range openAPIMethods {
			opRaw, ok := pathItem[method]
			if !ok {
				continue
			}
			op, _ := opRaw.(map[string]any)

			opID := stringFrontmatterOr(op["operationId"], "")
			if opID == "" {
				opID = fmt.Sprintf("%s_%s", method, strings.ReplaceAll(pathStr, "/", "_"))
			}
			conceptID := "api/" + openAPIOpIDCleanRE.ReplaceAllString(opID, "_")

			summary := stringFrontmatterOr(op["summary"], "")
			if summary == "" {
				summary = stringFrontmatterOr(op["description"], "")
			}
			if summary == "" {
				summary = fmt.Sprintf("%s %s", strings.ToUpper(method), pathStr)
			}
			description := stringFrontmatterOr(op["description"], summary)

			tags := append([]string{"api", method}, frontmatterStringSlice(op["tags"])...)

			concepts[conceptID] = Document{
				Frontmatter: map[string]any{
					"type":        "API",
					"title":       fmt.Sprintf("%s %s", strings.ToUpper(method), pathStr),
					"description": summary,
					"resource":    fmt.Sprintf("%s#%s", pathStr, method),
					"tags":        tags,
					"version":     version,
					"timestamp":   now,
				},
				Body: fmt.Sprintf(
					"# %s %s\n\n%s\n\n## Details\n- **Operation ID**: `%s`\n- **Method**: `%s`\n- **Path**: `%s`\n",
					strings.ToUpper(method), pathStr, description, opID, strings.ToUpper(method), pathStr,
				),
			}
		}
	}

	return concepts, nil
}

// ExportBundle fully owns the api/ namespace for this spec: re-harvesting
// after an operation was renamed or removed prunes its stale concept file.
func (e OpenAPIExtractor) ExportBundle(bundleRoot string) (int, error) {
	return exportBundle(bundleRoot, e.ExtractConcepts, []string{"api"})
}
