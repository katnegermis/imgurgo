package imgurgo

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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
	bResp, err := getBasicResponse(resp.Body)
	if err != nil {
		return nil, err
	} else if !bResp.Success {
		return nil, errors.New(fmt.Sprintf(
			"Failed to upload image. HTTP error code: %d", bResp.Status))
	}

	ui := bResp.getUploadedImage()
	ui.requester = r
	return ui, nil
}

// NewImageFromPath initializes an Image struct with raw data
// from the given image path. The user should initialize
// the rest of the fields in the struct, e.g. title, description etc.
func NewImageFromPath(imgPath string) (*Image, error) {
	f, err := os.Open(imgPath)
	if err != nil {
		return nil, err
	}
	var imgData bytes.Buffer
	io.Copy(&imgData, f)

	return &Image{Image: imgData.String(), Name: path.Base(imgPath)}, nil
}

// UploadedImage holds information about an image which resides on the servers of imgur.com.
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

	// Pointer to the requester which retrieved information about this image.
	// We know that the requester which retrieved the information is authorized
	// to modify it, and storing it in the image's struct let's us implement
	// functionality directly on the image, i.e. Delete, Update.
	requester *Requester
}

// GetId returns the id of an image; in case the image been uploaded anonymously
// it doesn't have an id, and the image's delete hash will be returned instead.
func (ui *UploadedImage) GetId() string {
	if len(ui.Id) <= 0 {
		return ui.DeleteHash
	}
	return ui.Id
}

func (ui *UploadedImage) Delete() error {
	delUrl := fmt.Sprintf("%s/%s", imgUrl, ui.GetId())
	resp, err := ui.requester.Do("DELETE", delUrl, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check whether request succeeded.
	bResp, err := getBasicResponse(resp.Body)
	if err != nil {
		return err
	}
	if !bResp.Success {
		return errors.New("Error deleting image")
	}

	// ui has been successfully deleted. Delete the reference to it
	// such that it no longer can be used.
	ui = nil
	return nil
}

func (ui *UploadedImage) UpdateTitleDesc(title, desc string) error {
	body := url.Values{"title": {title}, "description": {desc}}
	data := strings.NewReader(body.Encode())
	resp, err := ui.requester.Do("PUT", imgUrl, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check whether request succeeded.
	bResp, err := getBasicResponse(resp.Body)
	if err != nil {
		return err
	}
	if !bResp.Success {
		return errors.New("Error updating image data")
	}

	// Update succeeded, update local struct.
	ui.Title, ui.Description = title, desc
	return nil
}

func getBasicResponse(body io.Reader) (*basicResponse, error) {
	// Read response and decode JSON.
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	var bResp basicResponse
	if err := json.Unmarshal(b, &bResp); err != nil {
		return nil, err
	}
	return &bResp, nil
}
