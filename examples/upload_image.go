package main

import (
	"bufio"
	"fmt"
	"github.com/katnegermis/imgurgo"
	"log"
	"net/http"
	"os"
)

func main() {
	clientId := os.Getenv("IMGURGO_CLIENTID")
	clientSecret := os.Getenv("IMGURGO_CLIENTSECRET")
	path := "/home/katnegermis/pic.png"

	AnonymousExample(clientId, path)
	CodeExample(clientId, clientSecret, path)
	// PinExample(clientId, clientSecret, path)
}

func AnonymousExample(clientId, path string) {
	r := imgurgo.NewRequesterAnonymous(clientId)
	res, err := r.UploadImageFromPath(path)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(res)
}

func PinExample(clientId, clientSecret, path string) {
	r := imgurgo.NewRequesterAnonymous(clientId)

	// Wait for user to type PIN in to terminal.
	// The thing that really is relevant here, is the usage of SecretChan.
	go func() {
		r := bufio.NewReader(os.Stdin)
		fmt.Print("Please input PIN:")
		str, err := r.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		// Remove newline.
		r.Authorizer.SecretChan <- str[:len(str)-1]
	}()

	resp, err := r.UploadImageFromPath(path)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(resp)
}

func CodeExample(clientId, clientSecret, path string) {
	r := imgurgo.NewRequesterCode(clientId, clientSecret, "")

	// Start webserver to listen for imgur's callback.
	// The thing that really is relevant here, is the usage of SecretChan.
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			if err != nil {
				log.Fatal(err)
			}
			w.Write([]byte("You can now safely close this window."))
			r.Authorizer.SecretChan <- r.Form.Get("code")
		})
		http.ListenAndServe(":8080", nil)
	}()

	resp, err := r.UploadImageFromPath(path)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(resp)
}
