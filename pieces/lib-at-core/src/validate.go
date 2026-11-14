package atcore

import (
	"bytes"
	"encoding/json"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Validate checks v against the schema at path. v may hold ordinary Go values (ints, maps, slices):
// it is normalised through JSON so the validator sees json.Number, as it does for decoded files.
func (r *Registry) Validate(path string, v any) error {
	sch, err := r.Schema(path)
	if err != nil {
		return err
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	norm, err := jsonschema.UnmarshalJSON(bytes.NewReader(b))
	if err != nil {
		return err
	}
	return sch.Validate(norm)
}
