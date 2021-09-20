package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/hirochachacha/go-smb2"
)

type application struct {
	auth struct {
		username string
		password string
	}
	server struct {
		port       string
		datafolder string
		stage      string
	}
	smb struct {
		enabled    string
		servername string
		sharename  string
		username   string
		password   string
		domain     string
	}
}

func main() {

	app := new(application)
	app.auth.username = os.Getenv("AUTH_USERNAME")
	app.auth.password = os.Getenv("AUTH_PASSWORD")
	app.server.port = os.Getenv("SERVER_PORT")
	app.server.stage = os.Getenv("SERVER_STAGE")
	app.smb.enabled = os.Getenv("SMB_ENABLED")
	app.smb.servername = os.Getenv("SMB_SERVERNAME")
	app.smb.sharename = os.Getenv("SMB_SHARENAME")
	app.smb.username = os.Getenv("SMB_USERNAME")
	app.smb.password = os.Getenv("SMB_PASSWORD")
	app.smb.domain = os.Getenv("SMB_DOMAIN")

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

	if app.server.stage == "" {
		// Setting default Port
		app.server.stage = "prod"

	}

	if app.smb.enabled == "true" {
		if app.smb.servername == "" {
			log.Fatal("smb servername must be provided")
		}
		if app.smb.sharename == "" {
			log.Fatal("smb sharename must be provided")
		}
		if app.smb.username == "" {
			log.Fatal("smb username must be provided")
		}
		if app.smb.password == "" {
			log.Fatal("smb password must be provided")
		}
		if app.smb.domain == "" {
			log.Fatal("smb domain must be provided")
		}
	}

	if _, err := os.Stat("./data"); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir("./data", os.ModePerm)
		if err != nil {
			log.Fatal("Can't create data folder: ", err)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/"+app.server.stage+"-upload", app.basicAuth(app.getDataStream))

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

	nameOfFile, err := writeToFile(bodyBytes)
	if err != nil {
		log.Println(err)
	}

	if app.smb.enabled == "true" {
		err := pushOnSMB(nameOfFile, bodyBytes, app.smb.servername, app.smb.sharename, app.smb.username, app.smb.password, app.smb.domain)
		if err != nil {
			log.Println("SMB Upload failed!")
		}
		log.Printf("SMB: Successfully uploaded to \\\\%s\\%s\\%s", app.smb.servername, app.smb.sharename, nameOfFile)
	}

	fmt.Fprintf(w, "Successfully Uploaded File\n")
}

func writeToFile(b []byte) (string, error) {

	content := []byte(b)
	fileName := getFilenameDate()
	tempFile, err := ioutil.TempFile("data", fileName)
	if err != nil {
		log.Println(err)
		return "", err
	}

	if _, err = tempFile.Write(content); err != nil {
		log.Println("Failed to write the file", err)
	}
	if err := tempFile.Close(); err != nil {
		log.Println(err)
	}
	log.Printf("Local Storage: File written -> %+v\n", tempFile.Name())

	return fileName, nil
}

func getFilenameDate() string {
	const layout = "02-01-2006"
	t := time.Now()
	unix_timestamp := strconv.FormatInt(time.Now().UTC().Unix(), 10)
	return t.Format(layout) + "_" + unix_timestamp + ".cxml"
}

func pushOnSMB(filename string, fileContent []byte, servername string, sharename string, username string, password string, domain string) error {
	conn, err := net.Dial("tcp", servername+":445")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     username,
			Password: password,
			Domain:   domain,
		},
	}

	s, err := d.Dial(conn)
	if err != nil {
		panic(err)
	}
	defer s.Logoff()

	fs, err := s.Mount("\\\\" + servername + "\\" + sharename)
	if err != nil {
		panic(err)
	}
	defer fs.Umount()

	err = fs.WriteFile(filename, fileContent, 0444)
	if err != nil {
		log.Println("Couldn't write file to smb share", err)
	}

	return nil
}
