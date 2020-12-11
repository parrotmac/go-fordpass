package fordpass

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	fordAPIHost = "usapi.cv.ford.com"

	authHost = "fcis.ice.ibmcloud.com"
	authPath = "v1.0/endpoint/default/token"

	userAgent       = "fordpass-na/353 CFNetwork/1121.2.2 Darwin/19.3.0"
	jsonContentType = "application/json"
	formContentType = "application/x-www-form-urlencoded"
	applicationID   = "71A3AD0A-CF46-4CCF-B473-FC7FE5BC4592"
	authClientID    = "9fb503e0-715b-47e8-adfd-ad4b7770f73b"
	acceptLanguage  = "en-us"

	actionLockDoors   = "lock"
	actionUnlockDoors = "unlock"
	actionStartEngine = "start"
	actionStopEngine  = "stop"
)

var fordHeaders = map[string]string{
	"Accept":          "*/*",
	"Accept-Language": acceptLanguage,
	"Content-Type":    jsonContentType,
	"User-Agent":      userAgent,
	"Application-Id":  applicationID,
}

var iamHeaders = map[string]string{
	"Accept":          jsonContentType,
	"Accept-Language": acceptLanguage,
	"User-Agent":      userAgent,
	"Content-Type":    formContentType,
}

type authProvider struct {
	username  string
	password  string
	token     string
	expiresAt int64
}

func newAuthProvider(username, password string) authProvider {
	return authProvider{
		username:  username,
		password:  password,
		token:     "",
		expiresAt: 0,
	}
}

func (a *authProvider) authenticate(ctx context.Context) (AuthenticationResponse, error) {
	resp := AuthenticationResponse{}

	u := url.URL{
		Scheme: "https",
		Host:   authHost,
		Path:   authPath,
	}

	var formValues = map[string]string{
		"client_id":  "9fb503e0-715b-47e8-adfd-ad4b7770f73b",
		"grant_type": "password",
		"username":   a.username,
		"password":   a.password,
	}

	// Form
	v := url.Values{}
	for formKey, formValue := range formValues {
		v.Set(formKey, formValue)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), strings.NewReader(v.Encode()))
	if err != nil {
		return resp, err
	}

	for headerKey, headerValue := range iamHeaders {
		req.Header.Set(headerKey, headerValue)
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	httpResp, err := client.Do(req)
	if err != nil {
		return resp, err
	}

	bodyData, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return resp, err
	}

	return resp, json.Unmarshal(bodyData, &resp)
}

func (a *authProvider) GetAuthToken(ctx context.Context) (string, error) {
	if a.expiresAt <= time.Now().Add(5*time.Second).Unix() || a.token == "" {
		authResp, err := a.authenticate(ctx)
		if err != nil {
			return "", err
		}
		a.token = authResp.AccessToken
		a.expiresAt = time.Now().Unix() + int64(authResp.ExpiresIn)
	}
	return a.token, nil
}

type vehicleAPI struct {
	authProvider authProvider
	vin          string
}

func NewVehicleAPI(username, password, vin string) vehicleAPI {
	authProvider := newAuthProvider(username, password)
	return vehicleAPI{
		authProvider: authProvider,
		vin:          vin,
	}
}

func (a *vehicleAPI) Status(ctx context.Context) (VehicleStatusResponse, error) {
	resp := VehicleStatusResponse{}

	authToken, err := a.authProvider.GetAuthToken(ctx)
	if err != nil {
		return resp, err
	}

	urlQuery := url.Values{}
	urlQuery.Set("lrdt", "01-01-1970 00:00:00")

	u := url.URL{
		Scheme:   "https",
		Host:     fordAPIHost,
		Path:     fmt.Sprintf("api/vehicles/v4/%s/status", a.vin),
		RawQuery: urlQuery.Encode(),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return resp, err
	}

	for headerKey, headerValue := range fordHeaders {
		req.Header.Set(headerKey, headerValue)
	}
	req.Header.Set("auth-token", authToken)

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	httpResp, err := client.Do(req)
	if err != nil {
		return resp, err
	}

	bodyData, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return resp, err
	}

	return resp, json.Unmarshal(bodyData, &resp)
}

func (a *vehicleAPI) runVehicleAction(ctx context.Context, action string) (VehicleCommandResponse, error) {
	resp := VehicleCommandResponse{}

	authToken, err := a.authProvider.GetAuthToken(ctx)
	if err != nil {
		return resp, err
	}

	var urlPath string
	var httpMethod string
	switch action {
	case actionLockDoors:
		urlPath = fmt.Sprintf("api/vehicles/v2/%s/doors/lock", a.vin)
		httpMethod = http.MethodPut
	case actionUnlockDoors:
		urlPath = fmt.Sprintf("api/vehicles/v2/%s/doors/lock", a.vin)
		httpMethod = http.MethodDelete
	case actionStartEngine:
		urlPath = fmt.Sprintf("api/vehicles/v2/%s/engine/start", a.vin)
		httpMethod = http.MethodPut
	case actionStopEngine:
		urlPath = fmt.Sprintf("api/vehicles/v2/%s/engine/start", a.vin)
		httpMethod = http.MethodDelete
	default:
		return resp, fmt.Errorf("unrecognized action %+v", action)
	}

	u := url.URL{
		Scheme: "https",
		Host:   fordAPIHost,
		Path:   urlPath,
	}

	req, err := http.NewRequestWithContext(ctx, httpMethod, u.String(), nil)
	if err != nil {
		return resp, err
	}

	for headerKey, headerValue := range fordHeaders {
		req.Header.Set(headerKey, headerValue)
	}
	req.Header.Set("auth-token", authToken)

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	httpResp, err := client.Do(req)
	if err != nil {
		return resp, err
	}

	bodyData, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return resp, err
	}

	return resp, json.Unmarshal(bodyData, &resp)
}

func (a *vehicleAPI) inquireOnCommad(ctx context.Context, action string, commandID string) (int, error) {
	defaultResp := -1

	authToken, err := a.authProvider.GetAuthToken(ctx)
	if err != nil {
		return defaultResp, err
	}

	var urlPath string
	switch action {
	case actionLockDoors, actionUnlockDoors:
		urlPath = fmt.Sprintf("api/vehicles/v2/%s/doors/lock/%s", a.vin, commandID)
	case actionStartEngine, actionStopEngine:
		urlPath = fmt.Sprintf("api/vehicles/v2/%s/engine/start/%s", a.vin, commandID)
	default:
		return defaultResp, fmt.Errorf("unrecognized action %+v", action)
	}

	u := url.URL{
		Scheme: "https",
		Host:   fordAPIHost,
		Path:   urlPath,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return defaultResp, err
	}

	for headerKey, headerValue := range fordHeaders {
		req.Header.Set(headerKey, headerValue)
	}
	req.Header.Set("auth-token", authToken)

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	httpResp, err := client.Do(req)
	if err != nil {
		return defaultResp, err
	}

	bodyData, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return defaultResp, err
	}

	statusResp := VehicleCommandStatusResponse{}
	err = json.Unmarshal(bodyData, &statusResp)
	if err != nil {
		return defaultResp, err
	}
	return statusResp.Status, nil
}

func (a *vehicleAPI) Lock(ctx context.Context) error {
	return a.runCommand(ctx, actionLockDoors)
}

func (a *vehicleAPI) Unlock(ctx context.Context) error {
	return a.runCommand(ctx, actionUnlockDoors)
}

func (a *vehicleAPI) StartEngine(ctx context.Context) error {
	return a.runCommand(ctx, actionStartEngine)
}

func (a *vehicleAPI) StopEngine(ctx context.Context) error {
	return a.runCommand(ctx, actionStopEngine)
}

func (a *vehicleAPI) runCommand(ctx context.Context, action string) error {
	resp, err := a.runVehicleAction(ctx, action)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			commandStatus, err := a.inquireOnCommad(ctx, action, resp.CommandID)
			if err != nil {
				return err
			}

			switch commandStatus {
			case 200:
				// Success!
				return nil
			case 552:
				// Waiting for status -- check back later!
				continue
			default:
				return fmt.Errorf("Unrecognized command status %d", commandStatus)
			}
		}
	}
}
