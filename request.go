package imgurgo

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
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

func (r *Request) UploadImageFromPath(imgPath string) (*UploadedImage, error) {
	imgData, err := Base64EncodeFile(imgPath)
	if err != nil {
		return nil, err
	}
	image := &Image{Image: imgData.String(), Name: path.Base(imgPath)}
	return r.UploadImage(image)
}

func (r *Request) UploadImage(i *Image) (*UploadedImage, error) {
	if len(i.Image) == 0 {
		return nil, errors.New("Image.Image not set. There's nothing to upload.")
	}
	req, err := http.NewRequest("POST", imgUrl, strings.NewReader(i.encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if err := r.Authorizer.Authorize(req); err != nil {
		return nil, err
	}

	resp, err := r.do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Verify that response is ok.
	var bResp BasicResponse
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	json.Unmarshal(b, &bResp)
	if !bResp.Success {
		return nil, errors.New("Couldn't decode json response")
	}

	return bResp.getUploadedImage(), nil
}

func (r *Request) do(req *http.Request) (*http.Response, error) {
	// Make sure that r.client has been initialized.
	if r.client == nil {
		r.client = &http.Client{}
	}
	return r.client.Do(req)
}

func Base64EncodeFile(p string) (*bytes.Buffer, error) {
	var err error
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}

	var b64File bytes.Buffer
	// Is there a way to create a base64 encoder which doesn't require
	// us to copy the whole file to memory in order to decode? A base64
	// encoder which itself is a reader.
	b64Buf := base64.NewEncoder(base64.StdEncoding, &b64File)
	io.Copy(b64Buf, f)
	b64Buf.Close()
	return &b64File, nil
}
