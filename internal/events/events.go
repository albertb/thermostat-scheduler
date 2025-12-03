package events

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Specifies a peak demand event.
type PeakEvent struct {
	Start time.Time // Start of the peak demand event
	End   time.Time // End of the peak demand event
}

const winterPeakOfferURL = "https://donnees.hydroquebec.com/api/explore/v2.1/catalog/datasets/evenements-pointe/exports/json?lang=fr&refine=secteurclient%3A%22Residentiel%22&refine=offre%3A%22CPC-D%22&timezone=America%2FToronto"
const relevantOffer = "CPC-D"

func GetPeakEvents(cacheFilePath string, cacheTTL time.Duration, verbose bool) ([]PeakEvent, error) {
	offers, err := readCachedWinterPeakOffers(cacheFilePath, cacheTTL)
	if err == nil {
		return convertToPeakEvents(offers), nil
	}
	if verbose {
		log.Println("failed to read cached winter peak info:", err)
	}

	offers, err = fetchWinterPeakOffers(winterPeakOfferURL)
	if err != nil {
		return []PeakEvent{}, fmt.Errorf("failed to get winter peak info: %w", err)
	}

	err = writeCachedWinterPeakOffers(cacheFilePath, offers)
	if err != nil {
		log.Println("failed to write winter peak info cache:", err)
	}

	events := convertToPeakEvents(offers)
	for _, event := range events {
		if event.Start.After(time.Now()) {
			log.Println("upcoming peak event:", event)
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

func readCachedWinterPeakOffers(cacheFilePath string, cacheMaxAge time.Duration) ([]WinterPeakOffer, error) {
	var offers []WinterPeakOffer

	info, err := os.Stat(cacheFilePath)
	if err != nil {
		return offers, err
	}

	if time.Since(info.ModTime()) > cacheMaxAge {
		return offers, fmt.Errorf("cache is too old")
	}

	bytes, err := os.ReadFile(cacheFilePath)
	if err != nil {
		return offers, err
	}
	return parseWinterPeakOffers(bytes)
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

func writeCachedWinterPeakOffers(cacheFilePath string, offers []WinterPeakOffer) error {
	dir := filepath.Dir(cacheFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	bytes, err := json.Marshal(offers)
	if err != nil {
		return fmt.Errorf("failed to marshal winter peak info: %w", err)
	}

	if err := os.WriteFile(cacheFilePath, bytes, 0644); err != nil {
		return fmt.Errorf("failed to write winter peak info cache: %w", err)
	}
	return nil
}
