package imgurgo

import (
	"strconv"
	"time"
)

type basicResponse struct {
	Data    interface{}
	Success bool
	Status  int32
}

func (b *basicResponse) getUploadedImage() *UploadedImage {
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
		f, _ := strconv.ParseFloat(val.(string), 64)
		img.Size = f
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
