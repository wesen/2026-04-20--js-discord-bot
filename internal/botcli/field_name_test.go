package botcli

import "testing"

func TestRuntimeFieldInternalName(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"db-path", "db_path"},
		{"batch-size", "batch_size"},
		{"APIKey", "apikey"},
		{"simple", "simple"},
		{"dbPath", "db_path"},
		{"", ""},
		{"a-b-c", "a_b_c"},
		{"UPPER", "upper"},
	}

	for _, tc := range cases {
		got := runtimeFieldInternalName(tc.input)
		if got != tc.want {
			t.Errorf("runtimeFieldInternalName(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
