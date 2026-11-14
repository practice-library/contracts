package atcore

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// CodedError is an error that carries one of the repository's result codes.
type CodedError struct {
	Code string
	Msg  string
}

func (e *CodedError) Error() string { return e.Code + ": " + e.Msg }

// InputMissing is the code for a missing schema or policy file. It is a name, not a mirror of the
// code table, which Code checks at runtime.
const InputMissing = "commons.input_missing"

// readJSON reads a JSON file; numbers stay json.Number. A missing file is an InputMissing error.
func readJSON(path string) (any, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, &CodedError{InputMissing, path + " not found"}
	}
	if err != nil {
		return nil, err
	}
	v, err := jsonschema.UnmarshalJSON(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return v, nil
}

// LoadPolicy reads policy/commons.json under root.
func LoadPolicy(root string) (map[string]any, error) {
	v, err := readJSON(filepath.Join(root, "policy", "commons.json"))
	if err != nil {
		return nil, err
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("policy/commons.json: not an object")
	}
	return m, nil
}

// Registry holds every contract schema found under a root, identified by file path.
type Registry struct {
	root   string
	policy map[string]any
	c      *jsonschema.Compiler
	ids    map[string]string // slash path relative to root -> $id read from the file
}

// Load registers every pieces/contract-commons-*/schemas/*.schema.json under root, by the $id in the file.
func Load(root string, policy map[string]any) (*Registry, error) {
	r := &Registry{root: root, policy: policy, c: jsonschema.NewCompiler(), ids: map[string]string{}}
	r.c.DefaultDraft(jsonschema.Draft2020)
	files, err := filepath.Glob(filepath.Join(root, "pieces", "contract-commons-*", "schemas", "*.schema.json"))
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		v, err := readJSON(f)
		if err != nil {
			return nil, err
		}
		doc, _ := v.(map[string]any)
		id, _ := doc["$id"].(string)
		if id == "" {
			return nil, fmt.Errorf("%s: no $id", f)
		}
		if err := r.c.AddResource(id, v); err != nil {
			return nil, fmt.Errorf("%s: %w", f, err)
		}
		rel, _ := filepath.Rel(root, f)
		r.ids[filepath.ToSlash(rel)] = id
	}
	return r, nil
}

// Schema compiles the schema at path (relative to the root). An unknown path, or a referenced
// schema that was not loaded, is InputMissing.
func (r *Registry) Schema(path string) (*jsonschema.Schema, error) {
	id, ok := r.ids[path]
	if !ok {
		return nil, &CodedError{InputMissing, "schema " + path + " not found"}
	}
	sch, err := r.c.Compile(id)
	var lerr *jsonschema.LoadURLError
	if errors.As(err, &lerr) {
		return nil, &CodedError{InputMissing, "schema " + path + ": " + err.Error()}
	}
	return sch, err
}

// policyPath returns the string value of a policy key.
func (r *Registry) policyPath(key string) (string, error) {
	s, ok := r.policy[key].(string)
	if !ok || s == "" {
		return "", &CodedError{InputMissing, "policy key " + key + " not set"}
	}
	return s, nil
}
