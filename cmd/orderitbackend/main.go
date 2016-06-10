package main

import (
	"encoding/json"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
)

type receiveJSON struct {
	ownerName    string `json:"owner-name"`
	businessName string `json:"business-name"`
	email        string `json:"email"`
	password     string `json:"password"`
}

func register(rw http.ResponseWriter, req *http.Request) {
	var recv_json receiveJSON

	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(recv_json)
	if err != nil {
		log.Fatalf("Error decoding: %q", err)
		return
	}

	log.Println(string(recv_json.businessName))
	log.Println(recv_json.email)
	log.Println(recv_json.ownerName)
	log.Println(recv_json.password)

	encoder, err := json.Marshal(recv_json)
	if err != nil {
		log.Fatalf("Error marshal: %q", err)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.Write([]byte(encoder))
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/register", register)

	// cors.Default() setup the middleware with default options being
	// all origins accepted with simple methods (GET, POST). See
	// documentation below for more options.
	handler := cors.Default().Handler(mux)
	http.ListenAndServe(":"+port, handler)
}
