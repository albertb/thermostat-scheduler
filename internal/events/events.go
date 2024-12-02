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

const winterPeakOfferURL = "https://donnees.solutions.hydroquebec.com/donnees-ouvertes/data/json/pointeshivernales.json"
const relevantOffer = "CPC-D"

func GetPeakEvents(cacheFilePath string, cacheTTL time.Duration) ([]PeakEvent, error) {
	info, err := getCachedWinterPeakInfo(cacheFilePath, cacheTTL)
	if err == nil {
		return info.toPeakEvents(relevantOffer), nil
	}
	log.Println("failed to read cached winter peak info:", err)

	info, err = fetchWinterPeakInfo(winterPeakOfferURL)
	if err != nil {
		return []PeakEvent{}, err
	}

	err = writeWinterPeakInfoCache(cacheFilePath, info)
	if err != nil {
		log.Println("failed to write winter peak info cache:", err)
	}

	return info.toPeakEvents(relevantOffer), nil
}

type WinterPeakOffers struct {
	AvailableOffers []string `json:"offresDisponibles"`
	Events          []Event  `json:"evenements"`
}

type Event struct {
	Offer    string    `json:"offre"`         // Offers in effect during the event
	Start    time.Time `json:"dateDebut"`     // Start of the peak demand event
	End      time.Time `json:"dateFin"`       // End of the peak demand event
	Period   string    `json:"plageHoraire"`  // AM or PM
	Duration string    `json:"duree"`         // Duration using IOS8601, e.g., PT03H00MS for 3hr
	Sector   string    `json:"secteurClient"` // Résidentiel or Affaires
}

func (w WinterPeakOffers) toPeakEvents(offer string) []PeakEvent {
	var events []PeakEvent
	for _, e := range w.Events {
		if e.Offer != offer {
			continue
		}
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

func getCachedWinterPeakInfo(cacheFilePath string, cacheMaxAge time.Duration) (WinterPeakOffers, error) {
	info, err := os.Stat(cacheFilePath)
	if err != nil {
		return WinterPeakOffers{}, err
	}

	if time.Since(info.ModTime()) > cacheMaxAge {
		return WinterPeakOffers{}, fmt.Errorf("cache is too old")
	}

	bytes, err := os.ReadFile(cacheFilePath)
	if err != nil {
		return WinterPeakOffers{}, err
	}
	return parseWinterPeakInfo(bytes)
}

func fetchWinterPeakInfo(url string) (WinterPeakOffers, error) {
	resp, err := http.Get(url)
	if err != nil {
		return WinterPeakOffers{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return WinterPeakOffers{}, fmt.Errorf("HTTP request failed with: %s", resp.Status)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return WinterPeakOffers{}, err
	}

	var offers WinterPeakOffers
	if err := json.Unmarshal(bytes, &offers); err != nil {
		return offers, err
	}

	return offers, nil
}

func parseWinterPeakInfo(data []byte) (WinterPeakOffers, error) {
	var offers WinterPeakOffers
	err := json.Unmarshal(data, &offers)
	if err != nil {
		return WinterPeakOffers{}, err
	}
	return offers, nil
}

func writeWinterPeakInfoCache(cacheFilePath string, offers WinterPeakOffers) error {
	dir := filepath.Dir(cacheFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %v", err)
	}

	bytes, err := json.Marshal(offers)
	if err != nil {
		return err
	}

	return os.WriteFile(cacheFilePath, bytes, 0644)
}
