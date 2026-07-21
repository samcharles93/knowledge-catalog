package okf

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseConceptID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{name: "nested", input: "services/user-service", want: []string{"services", "user-service"}},
		{name: "surrounding slashes", input: "/services/users/", want: []string{"services", "users"}},
		{name: "underscore and dot", input: "api/v1_users.legacy", want: []string{"api", "v1_users.legacy"}},
		{name: "empty", input: "///", wantErr: true},
		{name: "traversal", input: "../outside", wantErr: true},
		{name: "space", input: "services/user service", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseConceptID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseConceptID(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseConceptID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestConceptPathRoundTrip(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	wantID := []string{"services", "user-service"}
	path, err := ConceptPath(root, wantID)
	if err != nil {
		t.Fatalf("ConceptPath() error = %v", err)
	}
	if want := filepath.Join(root, "services", "user-service.md"); path != want {
		t.Errorf("ConceptPath() = %q, want %q", path, want)
	}
	gotID, err := ConceptID(root, path)
	if err != nil {
		t.Fatalf("ConceptID() error = %v", err)
	}
	if !reflect.DeepEqual(gotID, wantID) {
		t.Errorf("ConceptID() = %v, want %v", gotID, wantID)
	}
}

func TestConceptIDRejectsPathOutsideBundle(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if _, err := ConceptID(root, filepath.Join(filepath.Dir(root), "outside.md")); err == nil {
		t.Fatal("ConceptID() error = nil, want outside-bundle error")
	}
}
