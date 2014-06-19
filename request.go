package imgurgo

import (
	"io"
	"net/http"
)

const (
	apiUrl    = "https://api.imgur.com/"
	apiVerUrl = apiUrl + "3/"
	imgUrl    = apiVerUrl + "image"
)

type Request struct {
	client     *http.Client
	Authorizer Authorizer
}

func NewRequest(authorizer Authorizer) *Request {
	return &Request{Authorizer: authorizer, client: &http.Client{}}
}

func (r *Request) UploadImageFromPath(path string) (*UploadedImage, error) {
	i, err := NewImageFromPath(path)
	if err != nil {
		return nil, err
	}
	return i.Upload(r)
}

func (r *Request) Do(method, url string, data io.Reader) (*http.Response, error) {
	// Make sure that r.client has been initialized.
	if r.client == nil {
		r.client = &http.Client{}
	}

	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if err := r.Authorizer.Authorize(req); err != nil {
		return nil, err
	}

	return r.client.Do(req)
}
