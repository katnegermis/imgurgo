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

type BasicResponse struct {
	Data    interface{}
	Success bool
	Status  int32
}

func (b *BasicResponse) getUploadedImage() *UploadedImage {
	img := &UploadedImage{}
	m := b.Data.(map[string]interface{})

	if val, ok := m["title"]; ok && val != nil {
		img.Title = val.(string)
	}
	if val, ok := m["type"]; ok && val != nil {
		img.Type = val.(string)
	}
	if val, ok := m["animated"]; ok && val != nil {
		img.Animated = val.(bool)
	}
	if val, ok := m["views"]; ok && val != nil {
		img.Views = val.(float64)
	}
	if val, ok := m["section"]; ok && val != nil {
		img.Section = val.(string)
	}
	if val, ok := m["description"]; ok && val != nil {
		img.Description = val.(string)
	}
	if val, ok := m["width"]; ok && val != nil {
		img.Width = val.(float64)
	}
	if val, ok := m["height"]; ok && val != nil {
		img.Height = val.(float64)
	}
	if val, ok := m["size"]; ok && val != nil {
		img.Size = val.(float64)
	}
	if val, ok := m["bandwidth"]; ok && val != nil {
		img.Bandwidth = val.(float64)
	}
	if val, ok := m["favorite"]; ok && val != nil {
		img.Favorite = val.(bool)
	}
	if val, ok := m["deletehash"]; ok && val != nil {
		img.DeleteHash = val.(string)
	}
	if val, ok := m["link"]; ok && val != nil {
		img.Link = val.(string)
	}
	if val, ok := m["datetime"]; ok && val != nil {
		img.DateTime = time.Unix(int64(val.(float64)), 0)
	}
	if val, ok := m["nsfw"]; ok && val != nil {
		img.NSFW = val.(bool)
	}
	return img
}

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
