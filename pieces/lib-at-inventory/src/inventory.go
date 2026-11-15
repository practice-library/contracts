// Package atinventory holds the inventory, affected-piece and case-file checks.
package atinventory

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"

	atcore "github.com/practice-library/contracts/pieces/lib-at-core/src"
)

// readJSON reads one JSON document; numbers stay json.Number. A missing file is InputMissing.
func readJSON(path string) (any, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, &atcore.CodedError{Code: atcore.InputMissing, Msg: path + " not found"}
	}
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, fmt.Errorf("%s: trailing data after the JSON value", path)
	}
	return v, nil
}

// Build validates every pieces/*/piece.json under root, sorted by path, and returns the inventory
// {"pieces":[{"path":...,"piece":...}]}, itself validated. Schemas are named in policy/commons.json.
func Build(root string, r *atcore.Registry) (map[string]any, error) {
	pol, err := atcore.LoadPolicy(root)
	if err != nil {
		return nil, err
	}
	pieceSchema, _ := pol["pieceSchema"].(string)
	invSchema, _ := pol["inventorySchema"].(string)
	files, err := filepath.Glob(filepath.Join(root, "pieces", "*", "piece.json"))
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	list := []any{}
	for _, f := range files {
		v, err := readJSON(f)
		if err != nil {
			return nil, err
		}
		if err := r.Validate(pieceSchema, v); err != nil {
			return nil, fmt.Errorf("%s: %w", f, err)
		}
		list = append(list, map[string]any{"path": "pieces/" + filepath.Base(filepath.Dir(f)), "piece": v})
	}
	inv := map[string]any{"pieces": list}
	if err := r.Validate(invSchema, inv); err != nil {
		return nil, fmt.Errorf("inventory: %w", err)
	}
	return inv, nil
}

// Canonical is inventory.json's form: sorted keys, indent 2, trailing newline.
func Canonical(v any) ([]byte, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	return append(b, '\n'), err
}

// Check compares Build's inventory with root inventory.json as JSON values, then byte for byte
// with Canonical. A missing file is InputMissing. This is the AT-05 partial stand-in.
func Check(root string, r *atcore.Registry) error {
	want, err := Build(root, r)
	if err != nil {
		return err
	}
	v, err := readJSON(filepath.Join(root, "inventory.json"))
	if err != nil {
		return err
	}
	have, _ := v.(map[string]any)
	hl, _ := have["pieces"].([]any)
	wl := want["pieces"].([]any)
	for i, w := range wl {
		if i >= len(hl) || !reflect.DeepEqual(w, hl[i]) {
			return fmt.Errorf("inventory.json is stale at %v", w.(map[string]any)["path"])
		}
	}
	if len(hl) > len(wl) {
		return fmt.Errorf("inventory.json is stale: extra entry %v", hl[len(wl)])
	}
	canon, _ := Canonical(want)
	if got, _ := os.ReadFile(filepath.Join(root, "inventory.json")); !bytes.Equal(got, canon) {
		return fmt.Errorf("inventory.json is not in canonical form (sorted keys, indent 2, trailing newline)")
	}
	return nil
}
