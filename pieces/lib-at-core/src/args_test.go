package atcore

import "testing"

func TestParse(t *testing.T) {
	good := "0123456789abcdef0123456789abcdef01234567"
	tests := []struct {
		name    string
		args    []string
		want    Args
		wantErr bool
	}{
		{"defaults", nil, Args{Root: "."}, false},
		{"all", []string{"--root", "/r", "--base", "main", "--commit", good}, Args{"/r", "main", good}, false},
		{"unknown flag", []string{"--nope"}, Args{}, true},
		{"positional", []string{"stray", "--commit", "zz"}, Args{}, true},
		{"short commit", []string{"--commit", "abc"}, Args{}, true},
		{"long commit", []string{"--commit", good + "0"}, Args{}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.args)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Fatalf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}
