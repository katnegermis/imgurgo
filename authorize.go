package imgurgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os/exec"
	"time"
)

const (
	oauthUrl = apiUrl + "oauth2/"
	authUrl  = oauthUrl + "authorize?client_id=%s&response_type=%s&state=%s"
	tokenUrl = oauthUrl + "token"

	ResponseTypePin   = "pin"
	ResponseTypeCode  = "code"
	ResponseTypeToken = "token"

	tokenLength = 40
)

type AuthResponse struct {
	AccessToken     string `json:"access_token"`
	RefreshToken    string `json:"refresh_token"`
	ExpirationTime  time.Time
	expiresIn       int32  `json:"expires_in"`
	TokenType       string `json:"token_type"`
	AccountUsername string `json:"account_username"`
}

type Anonymous struct {
	ClientId string
}

func (a *Anonymous) Authorize(req *http.Request) error {
	req.Header.Set("Authorization", fmt.Sprintf("Client-ID %s", a.ClientId))
	return nil
}

func NewAnonymousAuthorizer(clientId string) *Anonymous {
	return &Anonymous{ClientId: clientId}
}

type Pin struct {
	ClientId     string
	ClientSecret string
	State        string
	PinChan      chan<- string
	pinChan      chan string
	AuthData     *AuthResponse
}

func (p *Pin) Authorize(r *http.Request) error {
	oauthUrl := fmt.Sprintf(authUrl, p.ClientId, ResponseTypePin, p.State)

	// TODO: Make this work on all OSes...
	cmd := exec.Command("xdg-open", oauthUrl)
	err := cmd.Start()

	// Wait for token from client.
	pin := <-p.pinChan
	pin = pin[:len(pin)-1]

	// Trade pin for access token and refresh token.
	resp, err := http.PostForm(tokenUrl, url.Values{"client_id": {p.ClientId},
		"pin": {pin}, "client_secret": {p.ClientSecret}, "grant_type": {ResponseTypePin}})

	// Read and decode response
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var auth AuthResponse
	if err = json.Unmarshal(b, &auth); err != nil {
		return err
	}

	auth.ExpirationTime = time.Now().Add(time.Duration(auth.expiresIn) * time.Second)
	p.AuthData = &auth

	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", auth.AccessToken))

	return nil
}

func (p *Pin) RefreshAccessToken() error {
	if len(p.AuthData.RefreshToken) != tokenLength {
		return errors.New("Invalid refresh token!")
	}
	resp, err := http.PostForm(tokenUrl, url.Values{"client_id": {p.ClientId},
		"refresh_token": {p.AuthData.RefreshToken},
		"client_secret": {p.ClientSecret},
		"grant_type":    {ResponseTypePin}})
	if err != nil {
		return err
	}
	// Validate response, update data in p.
	resp = resp
	return errors.New("Not yet implemented.")
}

func NewPinAuthorizer(clientId, clientSecret, state string) *Pin {
	pinChan := make(chan string)
	return &Pin{ClientId: clientId, ClientSecret: clientSecret, State: state,
		pinChan: pinChan, PinChan: pinChan}
}
