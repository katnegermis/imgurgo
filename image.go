package imgurgo

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

type Image struct {
	// Image is a binary image file.
	Image       string
	Title       string
	Album       string
	Name        string
	Type        string
	Description string
}

func (i *Image) encode() string {
	val := url.Values{
		"image":       {i.Image},
		"title":       {i.Title},
		"album":       {i.Album},
		"name":        {i.Name},
		"type":        {i.Type},
		"description": {i.Description},
	}
	return val.Encode()
}

func (i *Image) Upload(r *Requester) (*UploadedImage, error) {
	if len(i.Image) == 0 {
		return nil, errors.New("Image.Image not set. There's nothing to upload.")
	}

	// Try to upload image
	resp, err := r.Do("POST", imgUrl, strings.NewReader(i.encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Verify that response is ok.
	var bResp basicResponse
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

func NewImageFromPath(imgPath string) (*Image, error) {
	f, err := os.Open(imgPath)
	if err != nil {
		return nil, err
	}
	var imgData bytes.Buffer
	io.Copy(&imgData, f)

	return &Image{Image: imgData.String(), Name: path.Base(imgPath)}, nil
}

type UploadedImage struct {
	Title       string
	Type        string
	Animated    bool
	Views       float64
	Section     string
	Description string
	Width       float64
	Height      float64
	Size        float64
	Bandwidth   float64
	Favorite    bool
	DeleteHash  string
	Link        string
	Id          string
	DateTime    time.Time
	NSFW        bool
}
