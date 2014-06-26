package imgurgo

import (
	"io"
	"net/http"
)

const (
	apiUrl    = "https://api.imgur.com/"
	apiVerUrl = apiUrl + "3/"
	imgUrl    = apiVerUrl + "image"
	userAgent = "imgurgo library"
)

type Requester struct {
	Authorizer Authorizer
}

func newRequester(authorizer Authorizer) *Requester {
	return &Requester{Authorizer: authorizer}
}

func NewRequesterAnonymous(clientId string) *Requester {
	a := newAnonymousAuthorizer(clientId)
	return newRequester(*a)
}

func NewRequesterPin(clientId, secret, state string) *Requester {
	a := newPinAuthorizer(clientId, secret, state)
	return newRequester(*a)
}

func NewRequesterCode(clientId, secret, state string) *Requester {
	a := newCodeAuthorizer(clientId, secret, state)
	return newRequester(*a)
}

func (r *Requester) UploadImageFromPath(path string) (*UploadedImage, error) {
	i, err := NewImageFromPath(path)
	if err != nil {
		return nil, err
	}
	return i.Upload(r)
}

func (r *Requester) Do(method, url string, data io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", userAgent)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Add authorization headers to the newly created request.
	if err := r.Authorizer.SetAuthHeaders(req); err != nil {
		return nil, err
	}

	return http.DefaultClient.Do(req)
}
