package rdap

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"

	"github.com/registrobr/rdap/Godeps/_workspace/src/github.com/aryann/difflib"
	"github.com/registrobr/rdap/Godeps/_workspace/src/github.com/davecgh/go-spew/spew"
)

func diff(a, b interface{}) []difflib.DiffRecord {
	return difflib.Diff(strings.Split(spew.Sdump(a), "\n"),
		strings.Split(spew.Sdump(b), "\n"))
}

func createTestServers(object interface{}, entry string) (*httptest.Server, *httptest.Server) {
	rdapTS := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(object)
		}),
	)

	return rdapTS, httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			registry := serviceRegistry{
				Version: "1.0",
				Services: []service{
					{
						{entry},
						{rdapTS.URL},
					},
				},
			}

			json.NewEncoder(w).Encode(registry)
		}),
	)
}

func objType(object interface{}) string {
	typ := reflect.TypeOf(object)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	return typ.Name()
}
