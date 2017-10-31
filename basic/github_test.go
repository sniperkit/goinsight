package basic

import (
	"testing"
)

func TestJoinMap(t *testing.T) {

	cases := []struct {
		input map[string]string
		want  string
	}{
		{map[string]string{"key": "value"}, "key:value"},
	}

	for _, c := range cases {
		got := GithubInsighter.joinMap(c.input)
		if got != c.want {
			t.Errorf("joinMap(%q) == %v, want %v", c.input, got, c.want)
		}
	}
}
