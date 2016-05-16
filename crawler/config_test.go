package crawler

import (
	"strings"
	"testing"
)

func TestReadConfig(t *testing.T) {
	jsonConfig := `
{
  "Default": {
    "TLS": true
  },
  "Sites": [
    {
      "Name": "foo"
    },
    {
      "Name": "bar",
      "TLS": false
    }
  ]
}
`
	cfg, err := readConfig(strings.NewReader(jsonConfig))
	if err != nil {
		t.Fatal(err)
	}
	var tests = []struct {
		i   int
		out bool
	}{
		{0, true},
		{1, false},
	}
	for _, tt := range tests {
		site := cfg.Sites[tt.i]
		if got := site.TLS; got != tt.out {
			t.Errorf("got TLS=%t, want TLS=%t for Name=%s", got, tt.out, site.Name)
		}
	}
}
