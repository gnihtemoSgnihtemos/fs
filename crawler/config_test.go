package crawler

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestReadConfig(t *testing.T) {
	jsonConfig := `
{
  "Database": "/tmp/foo.db",
  "Concurrency": 1,
  "Default": {
    "TLS": true,
    "ConnectTimeout": "1m",
    "ReadTimeout": "30s",
    "Ignore": [
      "foo"
    ]
  },
  "Sites": [
    {
      "Name": "foo",
      "ConnectTimeout": "10s",
      "ReadTimeout": "1m"
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
		i              int
		tls            bool
		ignore         []string
		connectTimeout time.Duration
		readTimeout    time.Duration
	}{
		{0, true, []string{"foo"}, 10 * time.Second, time.Minute},
		{1, false, []string{"bar"}, time.Minute, 30 * time.Second},
	}
	for _, tt := range tests {
		site := cfg.Sites[tt.i]
		if got := site.TLS; got != tt.tls {
			t.Errorf("got TLS=%t, want TLS=%t for Name=%s", got, tt.tls, site.Name)
		}
		if got := site.Ignore; !reflect.DeepEqual(got, tt.ignore) {
			t.Errorf("got Ignore=%s, want Ignore=%s for Name=%s", got, tt.ignore, site.Name)
		}
		if got := site.connectTimeout; site.connectTimeout != tt.connectTimeout {
			t.Errorf("got connectTimeout=%s, want connectTimeout=%s for Name=%s", got, tt.connectTimeout, site.Name)
		}
		if got := site.readTimeout; site.readTimeout != tt.readTimeout {
			t.Errorf("got readTimeout=%s, want readTimeout=%s for Name=%s", got, tt.readTimeout, site.Name)
		}
	}
}
