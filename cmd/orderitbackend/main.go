package main

import (
	"database/sql"
	"encoding/json"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var db *sql.DB = nil
var owner string = "owner"
var admin string = "admin"
var waiter string = "owner"
var staff string = "staff"

type jsonRecvOwner struct {
	OWNERNAME    string `json:"ownername"`
	BUSINESSNAME string `json:"businessname"`
	EMAIL        string `json:"email"`
	PASSWORD     string `json:"password"`
}

type jsonSendOwner struct {
	OWNERNAME    string `json:"ownername"`
	BUSINESSNAME string `json:"businessname"`
	EMAIL        string `json:"email"`
	PASSWORD     string `json:"password"`
	BUSINESSID   int64  `json:"businessid"`
}

type jsonRecvStaff struct {
	STAFFNAME     string `json:"staffName"`
	ROLEID        int64  `json:"roleId"`
	STAFFPASSWORD string `json:"staffPassword"`
	BUSINESSID    int64  `json:"businessId"`
}

func register(rw http.ResponseWriter, req *http.Request) {
	var recvjson = jsonRecvOwner{}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatalf("Error readall: %q", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &recvjson)
	if err != nil {
		log.Fatalf("Error decoding: %q", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert role if not exist
	var roleId int64 = 0
	err = db.QueryRow("SELECT id FROM roles WHERE name=$1", owner).Scan(&roleId)
	if err == sql.ErrNoRows {
		if err := db.QueryRow("INSERT INTO roles (name, type) VALUES ($1,$2) RETURNING id", owner, admin).Scan(&roleId); err != nil {
			log.Fatalf("Insert new role: %q", err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		log.Fatalf("Error selecting roles: %q", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert to businesses
	var businessId int64 = 0
	err = db.QueryRow("SELECT id FROM businesses WHERE name=$1", recvjson.BUSINESSNAME).Scan(&businessId)
	if err == sql.ErrNoRows {
		if err := db.QueryRow("INSERT INTO businesses (name, address) VALUES ($1,$2) RETURNING id", recvjson.BUSINESSNAME, "").Scan(&businessId); err != nil {
			log.Fatalf("Insert new businesses: %q", err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		log.Fatalf("Error selecting businesses: %q", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	} else {
		log.Fatal("Already exist business name")
		http.Error(rw, "Already exist business name", http.StatusInternalServerError)
		return
	}

	// Insert to users
	var userId int64 = 0
	err = db.QueryRow("SELECT id FROM users WHERE username=$1", recvjson.OWNERNAME).Scan(&userId)
	if err == sql.ErrNoRows {
		if err := db.QueryRow("INSERT INTO users (username, password, email, business_id, role_id) VALUES ($1,$2,$3,$4,$5) RETURNING id", recvjson.OWNERNAME, recvjson.PASSWORD, recvjson.EMAIL, businessId, roleId).Scan(&userId); err != nil {
			log.Fatalf("Insert new users: %q", err)
			return
		}
	} else if err != nil {
		log.Fatalf("Error selecting businesses: %q", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	} else {
		log.Fatal("Already exist owner name")
		http.Error(rw, "Already exist owner name", http.StatusInternalServerError)
		return
	}

	sendData := jsonSendOwner{
		OWNERNAME:    recvjson.OWNERNAME,
		BUSINESSNAME: recvjson.BUSINESSNAME,
		EMAIL:        recvjson.EMAIL,
		PASSWORD:     recvjson.PASSWORD,
		BUSINESSID:   businessId}

	encoder, err := json.Marshal(sendData)
	if err != nil {
		log.Fatalf("Error marshal: %q", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.Write(encoder)
}

func registerStaff(rw http.ResponseWriter, req *http.Request) {
	var jsonStaff = jsonRecvStaff{}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Fatalf("Error readall: %q", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &jsonStaff)
	if err != nil {
		log.Fatalf("Error decoding: %q", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert role if not exist
	var roleId int64 = 0
	err = db.QueryRow("SELECT id FROM roles WHERE id=$1", jsonStaff.ROLEID).Scan(&roleId)
	if err == sql.ErrNoRows {
		if err := db.QueryRow("INSERT INTO roles (name, type) VALUES ($1,$2) RETURNING id", waiter, staff).Scan(&roleId); err != nil {
			log.Fatalf("Insert new role: %q", err)
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		log.Fatalf("Error selecting roles: %q", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert to users
	var userId int64 = 0
	err = db.QueryRow("SELECT id FROM users WHERE username=$1", jsonStaff.STAFFNAME).Scan(&userId)
	if err == sql.ErrNoRows {
		if err := db.QueryRow("INSERT INTO users (username, password, email, business_id, role_id) VALUES ($1,$2,$3,$4,$5) RETURNING id", jsonStaff.STAFFNAME, jsonStaff.STAFFPASSWORD, "", jsonStaff.BUSINESSID, jsonStaff.ROLEID).Scan(&userId); err != nil {
			log.Fatalf("Insert new users: %q", err)
			return
		}
	} else if err != nil {
		log.Fatalf("Error selecting businesses: %q", err)
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	} else {
		log.Fatal("Already exist owner name")
		http.Error(rw, "Already exist owner name", http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte("true"))
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
		return
	}

	var err error

	db, err = sql.Open("postgres", "postgres://dsiedwaaywinyo:42TlG1tdtoxVYhKQDRX15mNaNs@ec2-54-163-239-28.compute-1.amazonaws.com:5432/d7mdp524vlsdf6?sslmode=require")
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
		return
	}

	// Create businessed table
	var BUSINESS_TABLE string = "CREATE TABLE IF NOT EXISTS businesses (id SERIAL PRIMARY KEY NOT NULL, name text, address text, created timestamp DEFAULT CURRENT_TIMESTAMP, modified timestamp DEFAULT CURRENT_TIMESTAMP)"
	if _, err = db.Exec(BUSINESS_TABLE); err != nil {
		log.Fatalf("Error creating businesses table: %q", err)
		return
	}

	// Create roles table
	var ROLE_TABLE string = "CREATE TABLE IF NOT EXISTS roles (id SERIAL PRIMARY KEY NOT NULL, name text, type text, created timestamp DEFAULT CURRENT_TIMESTAMP, modified timestamp DEFAULT CURRENT_TIMESTAMP)"
	if _, err = db.Exec(ROLE_TABLE); err != nil {
		log.Fatalf("Error creating roles table: %q", err)
		return
	}

	// Create users table
	var USER_TABLE string = "CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY NOT NULL, username text, password text, email text, business_id BIGINT references businesses(id), role_id BIGINT references roles(id), created timestamp DEFAULT CURRENT_TIMESTAMP, modified timestamp DEFAULT CURRENT_TIMESTAMP)"
	if _, err = db.Exec(USER_TABLE); err != nil {
		log.Fatalf("Error creating user table: %q", err)
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/register", register)
	mux.HandleFunc("/registerStaff", registerStaff)

	// cors.Default() setup the middleware with default options being
	// all origins accepted with simple methods (GET, POST). See
	// documentation below for more options.
	handler := cors.Default().Handler(mux)
	http.ListenAndServe(":"+port, handler)
}
