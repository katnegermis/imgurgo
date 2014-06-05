package imgurgo

import (
	"net/url"
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
