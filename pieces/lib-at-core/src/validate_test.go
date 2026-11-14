package atcore

import "testing"

func TestValidate(t *testing.T) {
	const xs = "pieces/contract-commons-x/schemas/x.schema.json"
	dir := fixture(t, map[string]string{
		xs: `{"$schema":"https://json-schema.org/draft/2020-12/schema","$id":"https://example.invalid/t/x.json",` +
			`"type":"object","required":["n"],"properties":{"n":{"type":"integer","minimum":0}}}`,
	})
	r, err := Load(dir, map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	for name, tc := range map[string]struct {
		v       any
		wantErr bool
	}{
		"go int":        {map[string]any{"n": 3}, false},
		"negative":      {map[string]any{"n": -1}, true},
		"missing field": {map[string]any{}, true},
		"wrong type":    {[]any{}, true},
	} {
		if err := r.Validate(xs, tc.v); (err != nil) != tc.wantErr {
			t.Errorf("%s: err = %v, wantErr %v", name, err, tc.wantErr)
		}
	}
	if err := r.Validate("pieces/contract-commons-x/schemas/none.schema.json", 1); !isMissing(err) {
		t.Fatalf("want input_missing, got %v", err)
	}
}
