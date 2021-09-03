package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type application struct {
	auth struct {
		username string
		password string
	}
	server struct {
		port       string
		datafolder string
	}
}

func main() {

	app := new(application)
	app.auth.username = os.Getenv("AUTH_USERNAME")
	app.auth.password = os.Getenv("AUTH_PASSWORD")
	app.server.port = os.Getenv("SERVER_PORT")

	if app.auth.username == "" {
		log.Fatal("basic auth username must be provided")
	}

	if app.auth.password == "" {
		log.Fatal("basic auth password must be provided")
	}

	if app.server.port == "" {
		// Setting default Port
		app.server.port = "80"
	}

	if _, err := os.Stat("./data"); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir("./data", os.ModePerm)
		if err != nil {
			log.Fatal("Can't create data folder: ", err)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/upload", app.basicAuth(app.getDataStream))

	srv := &http.Server{
		Addr:         ":" + app.server.port,
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Printf("starting post-handler on %s", srv.Addr)
	err := srv.ListenAndServe()
	log.Fatal(err)
}

func (app *application) basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if ok {
			usernameHash := sha256.Sum256([]byte(username))
			passwordHash := sha256.Sum256([]byte(password))
			expectedUsernameHash := sha256.Sum256([]byte(app.auth.username))
			expectedPasswordHash := sha256.Sum256([]byte(app.auth.password))

			usernameMatch := (subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:]) == 1)

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func (app *application) getDataStream(w http.ResponseWriter, r *http.Request) {

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}

	err = writeToFile(bodyBytes)
	if err != nil {
		log.Println(err)
	}

	fmt.Fprintf(w, "Successfully Uploaded File\n")
}

func writeToFile(b []byte) error {

	content := []byte(b)

	tempFile, err := ioutil.TempFile("data", "data-*.cxml")
	if err != nil {
		log.Println(err)
		return err
	}

	if _, err = tempFile.Write(content); err != nil {
		log.Println("Failed to write the file", err)
	}
	if err := tempFile.Close(); err != nil {
		log.Println(err)
	}
	fmt.Printf("File written: %+v\n", tempFile.Name())

	return nil
}
