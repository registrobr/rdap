package output

import (
	"fmt"
	"strings"

	"github.com/registrobr/rdap-client/Godeps/_workspace/src/github.com/aryann/difflib"
)

func diff(a, b interface{}) []difflib.DiffRecord {
	return difflib.Diff(strings.Split(fmt.Sprintf("%v", a), "\n"),
		strings.Split(fmt.Sprintf("%v", b), "\n"))
}
