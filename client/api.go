package client

import "fmt"

const UserAgent = "Braeburn/13 CFNetwork/1406.0.4 Darwin/22.4.0"

const BaseURL = "https://sd2.bluelinksmartconnect.com"
const LoginURL = BaseURL + "/api/v1/braeburn/rest-auth/login/"
const DevicesURL = BaseURL + "/api/v1/braeburn/devices/"
const deviceSetAttributeURLPattern = BaseURL + "/api/v1/braeburn/manage/%s/setstateattr/"

func DeviceSetAttributeURL(device string) string {
	return fmt.Sprintf(deviceSetAttributeURLPattern, device)
}

type LoginDetails struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthenticationKey struct {
	Key string `json:"key"`
}

type Device struct {
	UUID      string    `json:"uuid"` // The device identifier.
	StateData StateData `json:"state_data"`
}

type StateData struct {
	Program1 string `json:"PGM_01"` // Monday
	Program2 string `json:"PGM_02"` // Tuesday
	Program3 string `json:"PGM_03"` // Wednesday
	Program4 string `json:"PGM_04"` // Thursday
	Program5 string `json:"PGM_05"` // Friday
	Program6 string `json:"PGM_06"` // Saturday
	Program7 string `json:"PGM_07"` // Sunday
}
