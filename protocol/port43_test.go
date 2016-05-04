package protocol

import (
	"testing"
	"reflect"
)

func TestPort43SetPort43(t *testing.T) {
	expected := "whois.nic.br"

	var p Port43
	p.SetPort43(expected)

	if !reflect.DeepEqual(p.Port43, expected) {
		t.Errorf("Unexpected port43. Expected “%#v” and got “%#v”", expected, p.Port43)
	}
}
