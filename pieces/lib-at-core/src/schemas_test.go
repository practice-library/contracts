package atcore

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

const repoRoot = "../../.."

// fixture writes files (path -> content) under a new temp dir and returns it.
func fixture(t *testing.T, files map[string]string) string {
	dir := t.TempDir()
	for p, c := range files {
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

func realRegistry(t *testing.T) *Registry {
	p, err := LoadPolicy(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	r, err := Load(repoRoot, p)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func isMissing(err error) bool {
	var ce *CodedError
	return errors.As(err, &ce) && ce.Code == InputMissing
}

func TestLoadPolicy(t *testing.T) {
	if _, err := LoadPolicy(repoRoot); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadPolicy(t.TempDir()); !isMissing(err) {
		t.Fatalf("empty root: want input_missing, got %v", err)
	}
	for _, body := range []string{"[1]", "{"} {
		bad := fixture(t, map[string]string{"policy/commons.json": body})
		if _, err := LoadPolicy(bad); err == nil {
			t.Fatalf("policy %q accepted", body)
		}
	}
}

const (
	pkgReport = "pieces/contract-commons-at-report/schemas/"
	pkgPiece  = "pieces/contract-commons-piece/schemas/"
)

func TestLoadAndSchema(t *testing.T) {
	r := realRegistry(t)
	if _, err := r.Schema(pkgReport + "at-report.schema.json"); err != nil {
		t.Error(err)
	}
	if _, err := r.Schema("pieces/contract-none/schemas/x.schema.json"); !isMissing(err) {
		t.Errorf("unknown path: want input_missing, got %v", err)
	}
}

// A schema whose referenced schema was not loaded is input_missing, not an uncoded error.
func TestMissingReference(t *testing.T) {
	for _, tc := range []struct{ have, gone string }{
		{pkgReport + "at-report.schema.json", pkgReport + "codes.schema.json"},
		{pkgPiece + "inventory.schema.json", pkgPiece + "piece.schema.json"},
	} {
		b, err := os.ReadFile(filepath.Join(repoRoot, tc.have))
		if err != nil {
			t.Fatal(err)
		}
		r, err := Load(fixture(t, map[string]string{tc.have: string(b)}), map[string]any{})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := r.Schema(tc.have); !isMissing(err) {
			t.Errorf("%s without %s: want input_missing, got %v", tc.have, tc.gone, err)
		}
	}
}

func TestLoadFixture(t *testing.T) {
	const f = "pieces/contract-commons-a/schemas/x.schema.json"
	good := `{"$schema":"https://json-schema.org/draft/2020-12/schema","$id":"https://example.invalid/t/x.json"}`
	for body, wantErr := range map[string]bool{good: false, `{"type":"string"}`: true, "{": true} {
		if _, err := Load(fixture(t, map[string]string{f: body}), map[string]any{}); (err != nil) != wantErr {
			t.Errorf("%q: err = %v, wantErr %v", body, err, wantErr)
		}
	}
}
