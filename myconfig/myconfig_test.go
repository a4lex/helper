package myconfig

import (
	"bufio"
	"bytes"
	"testing"
)

const EXAMLE_CONFIG = `
#
# general configuration
#
general:
  log:
    file: ./my-program.log
    level: 63
#
# motion2rmq configuration
#
my-program:
  list:
    - firts_item
    - second_item
    - third_item
  rmq:
    host: localhost
    port: 5672
    user: guest
    password: guest
    queue: motion 
  log:
    file: /opt/var/log/my-program.log
`

func TestConfig(t *testing.T) {
	f := bytes.NewBufferString(EXAMLE_CONFIG)
	reader := bufio.NewReader(f)

	cfg, err := NewParser(reader).Parse()
	if err != nil {
		t.Error(err)
	}

	tests := []struct {
		key       string
		expectVal string
		err       error
	}{
		{"general.log.file", "./my-program.log", nil},
		{"general.log.level", "63", nil},
		{"my-program.rmq.host", "localhost", nil},
		{"my-program.rmq.port", "5672", nil},
		{"my-program.log.file", "/opt/var/log/my-program.log", nil},
		{"my-program.log.file.", "", ErrInvalidKey},
		{"my-program.log.files", "", ErrKeyNotFound},
	}

	for _, test := range tests {
		val, err := cfg.GetNodeValue(test.key)
		if err != nil {
			if err != test.err {
				t.Error(err)
			}
		} else if val != test.expectVal {
			t.Errorf("%s: expected %s, got %s", test.key, test.expectVal, val)
		}
	}

	list, err := cfg.GetNodeListValues("my-program.list")
	if err != nil {
		t.Error(err)
	}

	for i, item := range []string{"firts_item", "second_item", "third_item"} {
		if list[i] != item {
			t.Errorf("expected %s, got %s", item, list[i])
		}
	}
}
