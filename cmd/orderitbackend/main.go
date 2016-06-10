package main

import (
    "encoding/json"
    "log"
    "net/http"
)

type receiveJSON struct{
	ownerName  	string    `json:"owner-name"`
	businessName   string    `json:"business-name"`
	email    	string    `json:"email"`
	password     	string    `json:”password”`
}

func register(rw http.ResponseWriter, req *http.Request) {
    decoder := json.NewDecoder(req.Body)
    var recv_json receiveJSON   
    err := decoder.Decode(&recv_json)
    if err != nil {
        log.Fatalf("Error decoding: %q", err)
	return
    }
    log.Println(recv_json)
}

func main() {
	// Query string parameters are parsed using the existing underlying request object.
	// The request responds to a url matching:  /search?query=computer
	http.HandleFunc("/register", register)
}
