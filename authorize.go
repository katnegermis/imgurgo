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
	oAuthUrl     = apiUrl + "oauth2/"
	authUrl      = oAuthUrl + "authorize?client_id=%s&response_type=%s"
	authUrlState = oAuthUrl + "authorize?client_id=%s&response_type=%s&state=%s"
	tokenUrl     = oAuthUrl + "token"

	AuthTypeAnonymous AuthType = iota
	AuthTypePin
	AuthTypeCode

	responseTypePin   = "pin"
	responseTypeCode  = "code"
	responseTypeToken = "token"

	grantTypePin  = "pin"
	grantTypeCode = "authorization_code"

	tokenLength = 40
)

type Authorizer struct {
	AuthType     AuthType
	ClientId     string
	ClientSecret string
	RequestState string
	SecretChan   chan<- string
	secretChan   chan string
	AuthData     *AuthData

	responseType string
	grantType    string
}

func NewAnonymousAuthorizer(clientId string) *Authorizer {
	// Create AuthData in such a way that no OAuth requests will be sent.
	// This preserves the rest of the code from making special control
	// flow for the anonymous authorizer.
	authData := &AuthData{ExpirationTime: time.Now().Add(24 * 31 * time.Hour),
		AccessToken: clientId, TokenType: "Client-ID",
		RefreshToken: "0123456789012345678901234567890123456789"}
	return &Authorizer{ClientId: clientId, AuthType: AuthTypeAnonymous,
		AuthData: authData}
}

func NewPinAuthorizer(clientId, clientSecret, state string) *Authorizer {
	return newOAuthAuthorizer(AuthTypePin, clientId, clientSecret, state)
}

func NewCodeAuthorizer(clientId, clientSecret, state string) *Authorizer {
	return newOAuthAuthorizer(AuthTypeCode, clientId, clientSecret, state)
}

func (a *Authorizer) Authorize(r *http.Request) error {
	var err error
	// Check whether full authentication or refreshing of access token is needed.
	if !a.AccessTokenValid() {
		err = a.fullOAuthAuthentication()
	} else if a.AccessTokenExpired() && a.RefreshTokenValid() {
		err = a.RefreshAccessToken()
	}
	if err != nil {
		return err
	}

	r.Header.Set("Authorization", fmt.Sprintf("%s %s",
		a.AuthData.TokenType, a.AuthData.AccessToken))
	return nil
}

func (a *Authorizer) RefreshAccessToken() error {
	if !a.RefreshTokenValid() {
		return errors.New("Can't refresh access token: " +
			"Authorizer.AuthData.RefreshToken not valid.")
	}
	auth, err := postOAuthRequest(a.ClientId, a.ClientSecret,
		"refresh_token", "refresh_token", a.AuthData.RefreshToken)
	if err != nil {
		return err
	}
	a.AuthData = auth
	return nil
}

func (a *Authorizer) SetRefreshToken(token string) error {
	if len(token) != tokenLength {
		return errors.New(
			fmt.Sprintf("Invalid refresh token '%s' given!", token))
	}

	if a.AuthData == nil {
		a.AuthData = &AuthData{}
	}
	a.AuthData.RefreshToken = token
	return nil
}

func (a *Authorizer) AccessTokenExpired() bool {
	return time.Now().After(a.AuthData.ExpirationTime)
}

func (a *Authorizer) AccessTokenValid() bool {
	// Maybe we should try to actually authorize with imgur here, but that
	// would use the client's API credits. Instead we assume that the user
	// of the library wont set an invalid AccessToken himself.
	return a.AuthType == AuthTypeAnonymous ||
		(a.AuthData != nil && len(a.AuthData.AccessToken) == tokenLength)
}

func (a *Authorizer) RefreshTokenValid() bool {
	return a.AuthData != nil && len(a.AuthData.RefreshToken) == tokenLength
}

func (a *Authorizer) fullOAuthAuthentication() error {
	if a.AuthType == AuthTypeAnonymous {
		return errors.New("AuthTypeAnonymous can't do oAuth authentication!")
	}

	var oAuthUrl string
	if len(a.RequestState) == 0 {
		oAuthUrl = fmt.Sprintf(authUrl, a.ClientId, a.responseType)
	} else {
		oAuthUrl = fmt.Sprintf(authUrlState, a.ClientId, a.responseType,
			a.RequestState)
	}

	// Spawn browser to let user log in.
	// TODO: Make this platform independent.
	cmd := exec.Command("xdg-open", oAuthUrl)
	cmd.Start()

	// Wait for secret (pin or authorization code) which we need to trade for
	// our access- and refresh code.
	var secret string
	select {
	case <-time.Tick(1 * time.Minute):
		return errors.New(fmt.Sprintf(
			"We waited too long; the %s has timed out. Try again.", a.responseType))
	case secret = <-a.secretChan:
	}

	// Fetch AuthData from imgur.
	auth, err := postOAuthRequest(a.ClientId, a.ClientSecret,
		a.responseType, a.grantType, secret)
	if err != nil {
		return err
	}
	a.AuthData = auth

	return nil
}

type AuthType int8

type AuthData struct {
	AccessToken     string `json:"access_token"`
	RefreshToken    string `json:"refresh_token"`
	expiresIn       int32  `json:"expires_in"`
	TokenType       string `json:"token_type"`
	AccountUsername string `json:"account_username"`
	ExpirationTime  time.Time
}

func newOAuthAuthorizer(authType AuthType, clientId, clientSecret, state string) *Authorizer {
	var responseType, grantType string

	switch authType {
	case AuthTypeAnonymous:
		return NewAnonymousAuthorizer(clientId)

	case AuthTypePin:
		responseType = responseTypePin
		grantType = grantTypePin

	case AuthTypeCode:
		responseType = responseTypeCode
		grantType = grantTypeCode
	}

	secretChan := make(chan string)
	return &Authorizer{ClientId: clientId, ClientSecret: clientSecret,
		RequestState: state, SecretChan: secretChan, secretChan: secretChan,
		AuthType: authType, responseType: responseType, grantType: grantType}
}

func postOAuthRequest(clientId, clientSecret, responseType, grantType, secret string) (*AuthData, error) {
	resp, err := http.PostForm(tokenUrl, url.Values{"client_id": {clientId},
		responseType: {secret}, "client_secret": {clientSecret}, "grant_type": {grantType}})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and decode response
	b, err := ioutil.ReadAll(resp.Body)
	var auth AuthData
	if err = json.Unmarshal(b, &auth); err != nil {
		return nil, err
	}
	auth.ExpirationTime = time.Now().Add(time.Duration(auth.expiresIn) * time.Second)

	if auth.TokenType == "bearer" {
		auth.TokenType = "Bearer"
	}
	return &auth, nil
}
