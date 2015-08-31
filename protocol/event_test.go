package protocol

import (
	"reflect"
	"testing"
	"time"
)

func TestEventDateUnmarshalText(t *testing.T) {
	data := []struct {
		description string
		data        []byte
		expected    EventDate
	}{
		{
			description: "it should import a RFC3339 correctly",
			data:        []byte("2015-08-31T16:12:52Z"),
			expected: EventDate{
				Time: time.Date(2015, 8, 31, 16, 12, 52, 0, time.UTC),
			},
		},
		{
			description: "it should import a partial RFC3339 correctly",
			data:        []byte("2015-08-31"),
			expected: EventDate{
				Time: time.Date(2015, 8, 31, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			description: "it should fail for an invalid RFC3339",
			data:        []byte("31/8/2015"),
		},
	}

	for i, item := range data {
		var eventDate EventDate
		eventDate.UnmarshalText(item.data)

		if !reflect.DeepEqual(item.expected, eventDate) {
			t.Errorf("[%d] %s: unexpected event date returned. Expected “%#v” and got “%#v”", i, item.description, item.expected, eventDate)
		}
	}
}
