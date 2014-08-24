package main

import (
	"fmt"
	"github.com/katnegermis/imgurgo"
	"log"
	"os"
)

func main() {
	clientId := os.Getenv("IMGURGO_CLIENTID")
	r := imgurgo.NewRequesterAnonymous(clientId)
	for _, p := range os.Args[1:] {
		img, err := r.UploadImageFromPath(p)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Uploaded:\n    '%s'\nto\n    %s\n", p, img.Link)
	}
}
