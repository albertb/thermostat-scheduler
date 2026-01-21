package events

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Specifies a peak demand event.
type PeakEvent struct {
	Start time.Time // Start of the peak demand event
	End   time.Time // End of the peak demand event
}

const winterPeakOfferURL = "https://donnees.hydroquebec.com/api/explore/v2.1/catalog/datasets/evenements-pointe/exports/json?lang=fr&refine=secteurclient%3A%22Residentiel%22&refine=offre%3A%22CPC-D%22&timezone=America%2FToronto"

func GetPeakEvents(verbose bool) ([]PeakEvent, error) {
	offers, err := fetchWinterPeakOffers(winterPeakOfferURL)
	if err != nil {
		return []PeakEvent{}, fmt.Errorf("failed to get winter peak info: %w", err)
	}

	seenEvents, err := loadSeenEvents()
	if err != nil {
		return []PeakEvent{}, fmt.Errorf("failed to load seen events: %w", err)
	}

	events := convertToPeakEvents(offers)
	for _, event := range events {
		if event.Start.After(time.Now()) {
			if _, seen := seenEvents[eventID(event)]; !seen {
				log.Println("upcoming peak event:", event)
				if err := markEventAsSeen(event); err != nil {
					log.Printf("failed to mark event as seen: %v", err)
				}
			}
		}
	}
	return events, nil
}

type WinterPeakOffer struct {
	Offer    string    `json:"offre"`         // Offers in effect during the event
	Start    time.Time `json:"datedebut"`     // Start of the peak demand event
	End      time.Time `json:"datefin"`       // End of the peak demand event
	Period   string    `json:"plagehoraire"`  // AM or PM
	Duration string    `json:"duree"`         // Duration using IOS8601, e.g., PT03H00MS for 3hr
	Sector   string    `json:"secteurclient"` // Résidentiel or Affaires
}

func convertToPeakEvents(offers []WinterPeakOffer) []PeakEvent {
	var events []PeakEvent
	for _, e := range offers {
		if e.Start.After(e.End) {
			log.Println("Skipping invalid event:", e)
		}
		events = append(events, PeakEvent{
			Start: e.Start,
			End:   e.End,
		})
	}
	return events
}

func fetchWinterPeakOffers(url string) ([]WinterPeakOffer, error) {
	var offers []WinterPeakOffer

	resp, err := http.Get(url)
	if err != nil {
		return offers, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return offers, fmt.Errorf("HTTP request failed with: %s", resp.Status)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return offers, fmt.Errorf("failed to read HTTP response: %w", err)
	}
	return parseWinterPeakOffers(bytes)
}

func parseWinterPeakOffers(data []byte) ([]WinterPeakOffer, error) {
	var offers []WinterPeakOffer
	err := json.Unmarshal(data, &offers)
	if err != nil {
		return offers, fmt.Errorf("failed to parse winter peak info: %w", err)
	}
	return offers, nil
}
