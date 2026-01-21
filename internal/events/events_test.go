
package events

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetPeakEvents(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `[
			{
				"datedebut": "2024-01-24T06:00:00-05:00",
				"datefin": "2024-01-24T09:00:00-05:00",
				"offre": "CPC-D",
				"plagehoraire": "AM",
				"duree": "PT3H",
				"secteurclient": "Residentiel"
			}
		]`)
	}))
	defer server.Close()

	events, err := GetPeakEvents(server.URL, false)
	if err != nil {
		t.Fatalf("GetPeakEvents failed: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	expectedStart, _ := time.Parse(time.RFC3339, "2024-01-24T06:00:00-05:00")
	if events[0].Start != expectedStart {
		t.Errorf("expected start time %v, got %v", expectedStart, events[0].Start)
	}
}
