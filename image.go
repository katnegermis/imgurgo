package imgurgo

import (
	"net/url"
	"time"
)

type Image struct {
	// Image is a base64 encoded binary image file.
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
