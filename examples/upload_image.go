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
	// AnonymousExample(clientId, path)
	// PinExample(clientId, clientSecret, path)
	// CodeExample(clientId, clientSecret, path)
}

func AnonymousExample(clientId, path string) {
	a := imgurgo.NewAnonymousAuthorizer(clientId)
	r := &imgurgo.Request{Authorizer: *a}
	res, err := r.UploadImageFromPath(path)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(res)
}

func PinExample(clientId, clientSecret, path string) {
	a := imgurgo.NewPinAuthorizer(clientId, clientSecret, "")

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
		a.SecretChan <- str[:len(str)-1]
	}()

	r := imgurgo.NewRequest(*a)
	resp, err := r.UploadImageFromPath(path)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(resp)
}

func CodeExample(clientId, clientSecret, path string) {
	a := imgurgo.NewPinAuthorizer(clientId, clientSecret, "")

	// Start webserver to listen for imgur's callback.
	// The thing that really is relevant here, is the usage of SecretChan.
	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			if err != nil {
				log.Fatal(err)
			}
			w.Write([]byte("You can now safely close this window."))
			a.SecretChan <- r.Form.Get("code")
		})
		http.ListenAndServe(":8080", nil)
	}()

	r := imgurgo.NewRequest(*a)
	resp, err := r.UploadImageFromPath(path)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(resp)
}
