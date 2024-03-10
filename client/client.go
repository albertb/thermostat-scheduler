package client

import (
	"net/http"
	"net/http/cookiejar"
	"time"
)

type Client struct {
	client       *http.Client
	roundTripper *AuthorizingRoundTripper
}

func New() *Client {
	jar, _ := cookiejar.New(nil)
	rt := &AuthorizingRoundTripper{}
	return &Client{
		client: &http.Client{
			Jar:       jar,
			Timeout:   10 * time.Second,
			Transport: rt,
		}, roundTripper: rt}

}

type AuthorizingRoundTripper struct {
	Token string // The authentication string returned during login.
}

func (t *AuthorizingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(t.Token) > 0 {
		req.Header.Add("Authorization", t.Token)
	}
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return resp, err
	}
	return resp, err
}

func (c *Client) Login(username string, password string) error {
	loginDetails := LoginDetails{Username: username, Password: password}
	ak, err := Post[AuthenticationKey](c.client, LoginURL, loginDetails)
	if err != nil {
		return err
	}
	c.roundTripper.Token = "Token " + ak.Key
	return nil
}

func (c *Client) Devices() ([]Device, error) {
	var d []Device

	d, err := Get[[]Device](c.client, DevicesURL)
	if err != nil {
		return d, err
	}
	return d, nil
}

func (c *Client) SetDeviceAttributes(device string, data StateData) error {
	_, err := Post[Device](c.client, DeviceSetAttributeURL(device), data)
	if err != nil {
		return err
	}
	return nil
}
