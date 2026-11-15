package atcore

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
)

// Emit validates the report against the policy's reportSchema, then writes it as one JSON line.
func Emit(w io.Writer, r *Registry, report map[string]any) error {
	path, err := r.policyPath("reportSchema")
	if err != nil {
		return err
	}
	if err := r.Validate(path, report); err != nil {
		return err
	}
	b, err := json.Marshal(report)
	if err != nil {
		return err
	}
	_, err = w.Write(append(b, '\n'))
	return err
}

// Code returns name if it is in the $defs.code enum of the policy's codeTable, read at runtime.
func Code(r *Registry, name string) (string, error) {
	path, err := r.policyPath("codeTable")
	if err != nil {
		return "", err
	}
	v, err := readJSON(filepath.Join(r.root, path))
	if err != nil {
		return "", err
	}
	doc, _ := v.(map[string]any)
	defs, _ := doc["$defs"].(map[string]any)
	code, _ := defs["code"].(map[string]any)
	enum, _ := code["enum"].([]any)
	for _, e := range enum {
		if e == name {
			return name, nil
		}
	}
	return "", fmt.Errorf("unknown code %q", name)
}

// Exit is 0 when every report's status is PASS, else 1; an empty list is 1.
func Exit(reports []map[string]any) int {
	if len(reports) == 0 {
		return 1
	}
	for _, rep := range reports {
		if rep["status"] != "PASS" {
			return 1
		}
	}
	return 0
}
