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
	AuthTypeToken

	responseTypePin   = "pin"
	responseTypeCode  = "code"
	responseTypeToken = "token"

	grantTypePin  = "pin"
	grantTypeCode = "authorization_code"

	tokenLength = 40
)

type AuthType int8
type RefreshToken string

func (r RefreshToken) IsValid() bool {
	return len(r) == tokenLength
}

type AuthResponse struct {
	AccessToken     string       `json:"access_token"`
	RefreshToken    RefreshToken `json:"refresh_token"`
	expiresIn       int32        `json:"expires_in"`
	TokenType       string       `json:"token_type"`
	AccountUsername string       `json:"account_username"`
	ExpirationTime  time.Time
}

func NewAnonymousAuthorizer(clientId string) *Authorizer {
	return &Authorizer{ClientId: clientId, AuthType: AuthTypeAnonymous}
}

func (a *Authorizer) Authorize(r *http.Request) error {
	// Handle anonymous authentication.
	if a.AuthType == AuthTypeAnonymous {
		r.Header.Set("Authorization", fmt.Sprintf("Client-ID %s", a.ClientId))
		return nil
	}

	if a.AuthData == nil || !a.AuthData.RefreshToken.IsValid() {
		err := a.doOAuth()
		if err != nil {
			return err
		}
	} else if a.AuthData.ExpirationTime.After(time.Now()) && a.AuthData.RefreshToken.IsValid() {
		return a.RefreshAccessToken()
	}

	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", a.AuthData.AccessToken))
	return nil
}

func (a *Authorizer) RefreshAccessToken() error {
	return errors.New("Not yet implemented.")
}

func NewPinAuthorizer(clientId, clientSecret, state string) *Authorizer {
	return newAuthorizer(AuthTypePin, clientId, clientSecret, state)
}

func NewCodeAuthorizer(clientId, clientSecret, state string) *Authorizer {
	return newAuthorizer(AuthTypeCode, clientId, clientSecret, state)
}

func NewTokenAuthorizer(clientId, clientSecret, state string) *Authorizer {
	return newAuthorizer(AuthTypeToken, clientId, clientSecret, state)
}

func newAuthorizer(authType AuthType, clientId, clientSecret, state string) *Authorizer {
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

	case AuthTypeToken:
		responseType = responseTypeToken
	}

	secretChan := make(chan string)
	return &Authorizer{ClientId: clientId, ClientSecret: clientSecret, RequestState: state,
		SecretChan: secretChan, secretChan: secretChan, AuthType: authType,
		responseType: responseType, grantType: grantType}
}

func (a *Authorizer) doOAuth() error {
	if a.AuthType == AuthTypeAnonymous {
		return errors.New("AuthTypeAnonymous can't do oAuth authentication!")
	}

	var oAuthUrl string
	if len(a.RequestState) == 0 {
		oAuthUrl = fmt.Sprintf(authUrl, a.ClientId, a.responseType)
	} else {
		oAuthUrl = fmt.Sprintf(authUrlState, a.ClientId, a.responseType, a.RequestState)
	}

	// Spawn browser to let user log in
	cmd := exec.Command("xdg-open", oAuthUrl)
	// Start doesn't block, so I don't know for what/when we
	// could use the returned error.
	cmd.Start()

	// Wait for secret (pin or authorization code).
	secret := <-a.secretChan

	// This is a temporary hack to make token authentication work.
	// What we really want here, is to get the data that was sent to the
	// web server, i.e. access_token, token_type, and expires_in.
	if a.AuthType == AuthTypeToken {
		// TODO: figure out how we acquire AuthData from token request.
		a.AuthData.AccessToken = secret
		return nil
	}

	// Trade secret for access token and refresh token.
	// This is only required for PIN and code authentication.
	if a.AuthType == AuthTypePin || a.AuthType == AuthTypeCode {
		resp, err := http.PostForm(tokenUrl, url.Values{"client_id": {a.ClientId},
			a.responseType: {secret}, "client_secret": {a.ClientSecret}, "grant_type": {a.grantType}})
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Read and decode response
		b, err := ioutil.ReadAll(resp.Body)
		var auth AuthResponse
		if err = json.Unmarshal(b, &auth); err != nil {
			return err
		}
		auth.ExpirationTime = time.Now().Add(time.Duration(auth.expiresIn) * time.Second)
		a.AuthData = &auth
	}

	return nil
}
