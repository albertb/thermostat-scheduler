package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func toJSON(T any) ([]byte, error) {
	return json.Marshal(T)
}

func parseJSON[T any](s []byte) (T, error) {
	var r T
	if err := json.Unmarshal(s, &r); err != nil {
		return r, err
	}
	return r, nil
}

func Post[T any](client *http.Client, url string, data any) (T, error) {
	var m T
	b, err := toJSON(data)
	if err != nil {
		return m, err
	}

	byteReader := bytes.NewReader(b)
	r, err := http.NewRequest("POST", url, byteReader)
	if err != nil {
		return m, err
	}

	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Accept-Type", "application/json")
	r.Header.Add("User-Agent", UserAgent)

	res, err := client.Do(r)
	if err != nil {
		return m, err
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return m, err
	}

	if res.StatusCode != 200 {
		return m, errors.New(
			"Request failed: " + res.Status + "\n" + string(body))
	}
	return parseJSON[T](body)
}

func Get[T any](client *http.Client, url string) (T, error) {
	var m T

	r, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return m, err
	}

	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Accept-Type", "application/json")
	r.Header.Add("User-Agent", UserAgent)

	res, err := client.Do(r)
	if err != nil {
		return m, err
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return m, err
	}

	if res.StatusCode != 200 {
		return m, errors.New(
			"Request failed: " + res.Status + "\n" + string(body))
	}
	return parseJSON[T](body)
}
