package atcore

import (
	"bytes"
	"strings"
	"testing"
)

func report(status string, code any) map[string]any {
	return map[string]any{
		"at": "AT-01", "command": "commons-at run", "commit": strings.Repeat("a", 40),
		"status": status, "code": code, "counts": map[string]any{"checked": 2}, "findings": []any{},
	}
}

func TestEmit(t *testing.T) {
	r := realRegistry(t)
	bad := report("PASS", nil)
	delete(bad, "commit")
	tests := []struct {
		name    string
		rep     map[string]any
		wantErr bool
	}{
		{"pass", report("PASS", nil), false},
		{"not run with code", report("NOT_RUN", "commons.input_missing"), false},
		{"pass with code", report("PASS", "commons.input_missing"), true},
		{"missing field", bad, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := Emit(&buf, r, tc.rep)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr {
				if buf.Len() != 0 {
					t.Fatal("wrote output for an invalid report")
				}
				return
			}
			out := buf.String()
			if !strings.HasSuffix(out, "\n") || strings.Count(out, "\n") != 1 {
				t.Fatalf("want one line with a trailing newline, got %q", out)
			}
		})
	}
}

func TestEmitMissing(t *testing.T) {
	var buf bytes.Buffer
	for _, pol := range []map[string]any{
		{"reportSchema": "pieces/contract-none/schemas/r.schema.json"},
		{},
	} {
		r, err := Load(t.TempDir(), pol)
		if err != nil {
			t.Fatal(err)
		}
		if err := Emit(&buf, r, report("PASS", nil)); !isMissing(err) {
			t.Fatalf("policy %v: want input_missing, got %v", pol, err)
		}
	}
}

func TestCode(t *testing.T) {
	r := realRegistry(t)
	if got, err := Code(r, "commons.input_missing"); err != nil || got != "commons.input_missing" {
		t.Fatalf("known: got %q, %v", got, err)
	}
	if _, err := Code(r, "commons.made_up"); err == nil {
		t.Fatal("unknown code accepted")
	}
	for _, pol := range []map[string]any{{"codeTable": "pieces/contract-none/schemas/c.schema.json"}, {}} {
		r, _ = Load(t.TempDir(), pol)
		if _, err := Code(r, "commons.input_missing"); !isMissing(err) {
			t.Fatalf("policy %v: want input_missing, got %v", pol, err)
		}
	}
}

func TestExit(t *testing.T) {
	p, f, n := report("PASS", nil), report("FAIL", nil), report("NOT_RUN", "commons.input_missing")
	for want, in := range map[string][]map[string]any{
		"empty": nil, "all pass": {p, p}, "one fail": {p, f}, "one not run": {p, n},
	} {
		exp := 1
		if want == "all pass" {
			exp = 0
		}
		if got := Exit(in); got != exp {
			t.Errorf("%s: got %d, want %d", want, got, exp)
		}
	}
}
