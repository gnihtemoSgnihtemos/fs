package crawler

import (
	"reflect"
	"strings"
	"testing"
)

func TestReadConfig(t *testing.T) {
	jsonConfig := `
{
  "Default": {
    "TLS": true,
    "Ignore": [
      "foo"
    ]
  },
  "Sites": [
    {
      "Name": "foo"
    },
    {
      "Name": "bar",
      "TLS": false,
      "Ignore": [
        "bar"
      ]
    }
  ]
}
`
	cfg, err := readConfig(strings.NewReader(jsonConfig))
	if err != nil {
		t.Fatal(err)
	}
	var tests = []struct {
		i      int
		tls    bool
		ignore []string
	}{
		{0, true, []string{"foo"}},
		{1, false, []string{"bar"}},
	}
	for _, tt := range tests {
		site := cfg.Sites[tt.i]
		if got := site.TLS; got != tt.tls {
			t.Errorf("got TLS=%t, want TLS=%t for Name=%s", got, tt.tls, site.Name)
		}
		if got := site.Ignore; !reflect.DeepEqual(got, tt.ignore) {
			t.Errorf("got Ignore=%s, want Ignore=%s for Name=%s", got, tt.ignore, site.Name)
		}
	}
}
