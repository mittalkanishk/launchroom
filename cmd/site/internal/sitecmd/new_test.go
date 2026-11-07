package sitecmd

import "testing"

func TestParseNewArgs(t *testing.T) {
	cases := []struct {
		args     []string
		name     string
		template string
		ok       bool
	}{
		{[]string{"demo"}, "demo", "", true},
		{[]string{"demo", "--template", "astro"}, "demo", "astro", true},
		{[]string{"--template", "astro", "demo"}, "demo", "astro", true},
		{[]string{"--template=astro", "demo"}, "demo", "astro", true},
		{[]string{}, "", "", false},
		{[]string{"a", "b"}, "", "", false},
		{[]string{"demo", "--template"}, "", "", false},
		{[]string{"demo", "--force"}, "", "", false},
	}
	for _, c := range cases {
		name, template, ok := parseNewArgs(c.args)
		if name != c.name || template != c.template || ok != c.ok {
			t.Errorf("parseNewArgs(%q) = %q, %q, %v; want %q, %q, %v", c.args, name, template, ok, c.name, c.template, c.ok)
		}
	}
}
