package atinventory

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"
)

// inventory decodes an inventory document.
func inventory(t *testing.T, doc string) map[string]any {
	var inv map[string]any
	if err := json.Unmarshal([]byte(doc), &inv); err != nil {
		t.Fatal(err)
	}
	return inv
}

// entry is one inventory entry with the three consumes lists, each given as comma-joined strings.
func entry(path, consumes, testConsumes, generatorConsumes string) string {
	return `{"path":"` + path + `","piece":{"consumes":[` + consumes + `],"testConsumes":[` + testConsumes +
		`],"generatorConsumes":[` + generatorConsumes + `]}}`
}

// graph: b consumes a; c test-consumes b; d generator-consumes a; e consumes only a schema id;
// f stands alone.
var graph = `{"pieces":[` + strings.Join([]string{
	entry("pieces/a", ``, ``, ``),
	entry("pieces/b", `"pieces/a"`, ``, ``),
	entry("pieces/c", ``, `"pieces/b"`, ``),
	entry("pieces/d", ``, ``, `"pieces/a"`),
	entry("pieces/e", `"example.invalid:schema:1.0.0"`, ``, ``),
	entry("pieces/f", ``, ``, ``),
}, ",") + `]}`

func TestAffected(t *testing.T) {
	all := []string{"pieces/a", "pieces/b", "pieces/c", "pieces/d", "pieces/e", "pieces/f"}
	for name, tc := range map[string]struct {
		changed []string
		want    []string
	}{
		"no change":                {nil, []string{}},
		"root file":                {[]string{"go.mod"}, all},
		"go.sum":                   {[]string{"go.sum"}, all},
		"workflow":                 {[]string{".github/workflows/gate.yml"}, all},
		"policy":                   {[]string{"policy/commons.json"}, all},
		"inventory":                {[]string{"inventory.json"}, all},
		"file in pieces":           {[]string{"pieces/README.md"}, all},
		"unlisted piece directory": {[]string{"pieces/g/src/x.go"}, all},
		"name prefix of a piece":   {[]string{"pieces/a-extra/x"}, all},
		"contract change":          {[]string{"pieces/a/schemas/x.schema.json"}, []string{"pieces/a", "pieces/b", "pieces/c", "pieces/d"}},
		"the piece directory":      {[]string{"pieces/a"}, []string{"pieces/a", "pieces/b", "pieces/c", "pieces/d"}},
		"middle of the chain":      {[]string{"pieces/b/README.md"}, []string{"pieces/b", "pieces/c"}},
		"leaf consumer":            {[]string{"pieces/c/src/x.go"}, []string{"pieces/c"}},
		"schema id is no edge":     {[]string{"pieces/e/piece.json"}, []string{"pieces/e"}},
		"standalone":               {[]string{"pieces/f/x"}, []string{"pieces/f"}},
		"duplicates and order":     {[]string{"pieces/c/x", "pieces/b/y", "pieces/c/z"}, []string{"pieces/b", "pieces/c"}},
		"dot segments":             {[]string{"./pieces/f/../f/x"}, []string{"pieces/f"}},
		"one outside path":         {[]string{"pieces/f/x", "go.mod"}, all},
	} {
		got, err := Affected(inventory(t, graph), tc.changed)
		if err != nil || !slices.Equal(got, tc.want) {
			t.Errorf("%s: got %v, %v; want %v", name, got, err, tc.want)
		}
	}
}

func TestAffectedErrors(t *testing.T) {
	for name, tc := range map[string]struct {
		doc     string
		wantErr string
	}{
		"cycle": {`{"pieces":[` + entry("pieces/a", `"pieces/b"`, ``, ``) + `,` +
			entry("pieces/b", ``, ``, `"pieces/a"`) + `]}`, "cycle"},
		"self cycle": {`{"pieces":[` + entry("pieces/a", ``, `"pieces/a"`, ``) + `]}`, "cycle"},
		"long cycle": {`{"pieces":[` + entry("pieces/a", `"pieces/b"`, ``, ``) + `,` +
			entry("pieces/b", `"pieces/c"`, ``, ``) + `,` + entry("pieces/c", ``, `"pieces/a"`, ``) + `]}`, "cycle"},
		"unknown piece": {`{"pieces":[` + entry("pieces/a", `"pieces/x"`, ``, ``) + `]}`, "pieces/x"},
		"no pieces":     {`{}`, "no pieces"},
		"bad entry":     {`{"pieces":[{"path":"pieces/a"}]}`, "malformed"},
	} {
		_, err := Affected(inventory(t, tc.doc), []string{"pieces/a/x"})
		if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
			t.Errorf("%s: want error with %q, got %v", name, tc.wantErr, err)
		}
	}
}

func TestAffectedRealInventory(t *testing.T) {
	inv := realInventory(t)
	got, err := Affected(inv, []string{"pieces/contract-commons-piece/schemas/piece.schema.json"})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(got, "pieces/lib-at-core") || !slices.Contains(got, "pieces/lib-at-inventory") {
		t.Fatalf("a contract change must reach its consumers, got %v", got)
	}
	if got, _ := Affected(inv, []string{"go.mod"}); len(got) < 4 {
		t.Fatalf("a root file must affect every piece, got %v", got)
	}
}
