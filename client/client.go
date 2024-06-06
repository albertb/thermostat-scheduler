package client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"thermostat-scheduler/api"
	"time"
)

type Client struct {
	baseURL   *url.URL
	userAgent string

	httpClient   *http.Client
	roundTripper *AuthorizingRoundTripper
}

func New() *Client {
	baseURL, _ := url.Parse("https://sd2.bluelinksmartconnect.com/api/v1/braeburn/")
	jar, _ := cookiejar.New(nil)

	rt := &AuthorizingRoundTripper{}
	return &Client{
		baseURL:   baseURL,
		userAgent: "Braeburn/13 CFNetwork/1406.0.4 Darwin/22.4.0",
		httpClient: &http.Client{
			Jar:       jar,
			Timeout:   10 * time.Second,
			Transport: rt,
		}, roundTripper: rt}

}

type AuthorizingRoundTripper struct {
	token *string // The authentication string returned during login.
}

func (t *AuthorizingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.token != nil {
		req.Header.Add("Authorization", *t.token)
	}
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return resp, err
	}
	return resp, err
}

func (c *Client) Login(username string, password string) error {
	loginDetails := api.LoginDetails{Username: username, Password: password}

	req, err := c.newRequest("POST", "rest-auth/login/", loginDetails)
	if err != nil {
		return err
	}

	var auth api.AuthenticationKey
	_, err = c.do(req, &auth)
	if err != nil {
		return err
	}

	token := "Token " + auth.Key
	c.roundTripper.token = &token
	return nil
}

func (c *Client) Devices() ([]api.Device, error) {
	req, err := c.newRequest("GET", "devices/", nil)
	if err != nil {
		return nil, err
	}

	var devices []api.Device
	_, err = c.do(req, &devices)
	if err != nil {
		return nil, err
	}
	return devices, nil
}

func (c *Client) SetDeviceAttributes(deviceId string, data api.StateData) (api.Device, error) {
	req, err := c.newRequest("POST", "manage/"+deviceId+"/setstateattr/", data)

	if err != nil {
		return api.Device{}, err
	}

	var updated api.Device
	_, err = c.do(req, &updated)
	if err != nil {
		return api.Device{}, err
	}

	return updated, nil
}

func (c *Client) newRequest(method, path string, body any) (*http.Request, error) {
	relative := &url.URL{Path: path}
	url := c.baseURL.ResolveReference(relative)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url.String(), buf)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	return req, nil
}

func (c *Client) do(req *http.Request, v any) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(v)
	return resp, err
}
