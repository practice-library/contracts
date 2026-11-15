package atinventory

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	atcore "github.com/practice-library/contracts/pieces/lib-at-core/src"
)

const repoRoot = "../../.."

// fixture writes files (path -> content) under a new temp dir and returns it. An empty content
// removes the path from the set, so a caller can override or drop a default file.
func fixture(t *testing.T, base, extra map[string]string) string {
	dir := t.TempDir()
	files := map[string]string{}
	for _, m := range []map[string]string{base, extra} {
		for p, c := range m {
			files[p] = c
		}
	}
	for p, c := range files {
		if c == "" {
			continue
		}
		full := filepath.Join(dir, p)
		err := os.MkdirAll(filepath.Dir(full), 0o755)
		if err == nil {
			err = os.WriteFile(full, []byte(c), 0o644)
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func registry(t *testing.T, root string) *atcore.Registry {
	r, err := atcore.Load(root, map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func realRegistry(t *testing.T) *atcore.Registry {
	p, err := atcore.LoadPolicy(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	r, err := atcore.Load(repoRoot, p)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func isMissing(err error) bool {
	var ce *atcore.CodedError
	return errors.As(err, &ce) && ce.Code == atcore.InputMissing
}

const (
	schemaDir = "pieces/contract-commons-piece/schemas/"
	draft     = `"$schema":"https://json-schema.org/draft/2020-12/schema",`
)

// invFiles is a small repository: the piece and inventory schemas, the policy naming them, and
// two pieces where b consumes a.
var invFiles = map[string]string{
	schemaDir + "piece.schema.json": `{` + draft + `"$id":"https://example.invalid/t/piece.json","type":"object",` +
		`"required":["consumes"],"properties":{"consumes":{"type":"array","items":{"type":"string"}}}}`,
	schemaDir + "inventory.schema.json": `{` + draft + `"$id":"https://example.invalid/t/inventory.json","type":"object",` +
		`"required":["pieces"],"properties":{"pieces":{"type":"array","items":{"type":"object",` +
		`"required":["path","piece"],"properties":{"path":{"type":"string"},` +
		`"piece":{"$ref":"https://example.invalid/t/piece.json"}}}}}}`,
	"policy/commons.json": `{"pieceSchema":"` + schemaDir + `piece.schema.json",` +
		`"inventorySchema":"` + schemaDir + `inventory.schema.json"}`,
	"pieces/a/piece.json": `{"consumes":[]}`,
	"pieces/b/piece.json": `{"consumes":["pieces/a"]}`,
}

// writeInventory generates inventory.json in dir from its pieces, as the generator does.
func writeInventory(t *testing.T, dir string) {
	inv, err := Build(dir, registry(t, dir))
	if err == nil {
		var b []byte
		if b, err = Canonical(inv); err == nil {
			err = os.WriteFile(filepath.Join(dir, "inventory.json"), b, 0o644)
		}
	}
	if err != nil {
		t.Fatal(err)
	}
}

func TestRealInventory(t *testing.T) {
	r := realRegistry(t)
	if err := Check(repoRoot, r); err != nil {
		t.Fatal(err)
	}
	inv, err := Build(repoRoot, r)
	if err != nil {
		t.Fatal(err)
	}
	if list, _ := inv["pieces"].([]any); len(list) < 4 {
		t.Fatalf("want every real piece, got %d", len(list))
	}
}

func TestBuild(t *testing.T) {
	for name, tc := range map[string]struct {
		extra   map[string]string
		wantErr string
	}{
		"two pieces":       {nil, ""},
		"piece fails":      {map[string]string{"pieces/c/piece.json": `{}`}, "pieces/c/piece.json"},
		"piece not json":   {map[string]string{"pieces/c/piece.json": `{`}, "pieces/c/piece.json"},
		"trailing value":   {map[string]string{"pieces/c/piece.json": `{"consumes":[]} {}`}, "trailing data"},
		"trailing garbage": {map[string]string{"pieces/c/piece.json": `{"consumes":[]} x`}, "trailing data"},
		"no policy":        {map[string]string{"policy/commons.json": ""}, "input_missing"},
		"no piece schema":  {map[string]string{schemaDir + "piece.schema.json": ""}, "input_missing"},
		"unlisted policy":  {map[string]string{"policy/commons.json": `{}`}, "input_missing"},
		"no pieces at all": {map[string]string{"pieces/a/piece.json": "", "pieces/b/piece.json": ""}, ""},
	} {
		dir := fixture(t, invFiles, tc.extra)
		inv, err := Build(dir, registry(t, dir))
		switch {
		case tc.wantErr == "" && err != nil:
			t.Errorf("%s: %v", name, err)
		case tc.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)):
			t.Errorf("%s: want error with %q, got %v", name, tc.wantErr, err)
		case err == nil && name == "two pieces":
			list, _ := inv["pieces"].([]any)
			first, _ := list[0].(map[string]any)
			if len(list) != 2 || first["path"] != "pieces/a" {
				t.Errorf("%s: pieces not sorted by path: %v", name, list)
			}
		}
	}
}

func TestCheck(t *testing.T) {
	for name, tc := range map[string]struct {
		edit    func(t *testing.T, dir string)
		wantErr string // "" passes; "input_missing" is the code
	}{
		"fresh": {func(t *testing.T, dir string) {}, ""},
		"missing inventory": {func(t *testing.T, dir string) {
			os.Remove(filepath.Join(dir, "inventory.json"))
		}, "input_missing"},
		"stale piece": {func(t *testing.T, dir string) {
			os.WriteFile(filepath.Join(dir, "pieces/b/piece.json"), []byte(`{"consumes":[]}`), 0o644)
		}, "stale at pieces/b"},
		"new piece": {func(t *testing.T, dir string) {
			os.MkdirAll(filepath.Join(dir, "pieces/c"), 0o755)
			os.WriteFile(filepath.Join(dir, "pieces/c/piece.json"), []byte(`{"consumes":[]}`), 0o644)
		}, "stale at pieces/c"},
		"removed piece": {func(t *testing.T, dir string) {
			os.Remove(filepath.Join(dir, "pieces/b/piece.json"))
		}, "extra entry"},
		"invalid piece": {func(t *testing.T, dir string) {
			os.WriteFile(filepath.Join(dir, "pieces/a/piece.json"), []byte(`{"kind":"x"}`), 0o644)
		}, "pieces/a/piece.json"},
		"compact bytes": {func(t *testing.T, dir string) {
			os.WriteFile(filepath.Join(dir, "inventory.json"),
				[]byte(`{"pieces":[{"path":"pieces/a","piece":{"consumes":[]}},{"path":"pieces/b","piece":{"consumes":["pieces/a"]}}]}`),
				0o644)
		}, "canonical form"},
		"no trailing newline": {func(t *testing.T, dir string) {
			b, _ := os.ReadFile(filepath.Join(dir, "inventory.json"))
			os.WriteFile(filepath.Join(dir, "inventory.json"), []byte(strings.TrimSuffix(string(b), "\n")), 0o644)
		}, "canonical form"},
		"inventory not json": {func(t *testing.T, dir string) {
			os.WriteFile(filepath.Join(dir, "inventory.json"), []byte(`{`), 0o644)
		}, "inventory.json"},
		"inventory not an inventory": {func(t *testing.T, dir string) {
			os.WriteFile(filepath.Join(dir, "inventory.json"), []byte(`[]`), 0o644)
		}, "stale at pieces/a"},
	} {
		dir := fixture(t, invFiles, nil)
		writeInventory(t, dir)
		tc.edit(t, dir)
		err := Check(dir, registry(t, dir))
		switch {
		case tc.wantErr == "" && err != nil:
			t.Errorf("%s: %v", name, err)
		case tc.wantErr == "input_missing" && !isMissing(err):
			t.Errorf("%s: want input_missing, got %v", name, err)
		case tc.wantErr != "" && tc.wantErr != "input_missing" && (err == nil || !strings.Contains(err.Error(), tc.wantErr)):
			t.Errorf("%s: want error with %q, got %v", name, tc.wantErr, err)
		}
	}
}

func realInventory(t *testing.T) map[string]any {
	inv, err := Build(repoRoot, realRegistry(t))
	if err != nil {
		t.Fatal(err)
	}
	return inv
}
