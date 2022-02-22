package util

import (
	"errors"
	"os"
	"strconv"
)

type Application struct {
	Auth struct {
		Username string
		Password string
	}
	Server struct {
		Port       string
		Datafolder string
	}
	SMB struct {
		Enabled    bool
		Servername string
		Sharename  string
		Username   string
		Password   string
		Domain     string
	}
}

var (
	err error        = nil
	App *Application = &Application{}
)

// LoadEnv loads OS environment variables
func LoadEnv() error {

	if App.Auth.Username = os.Getenv("AUTH_USERNAME"); App.Auth.Username == "" {
		err = errors.New("basic auth username must be provided")
		return err
	}

	if App.Auth.Password = os.Getenv("AUTH_PASSWORD"); App.Auth.Password == "" {
		err = errors.New("basic auth password must be provided")
		return err
	}

	if App.Server.Port = os.Getenv("SERVER_PORT"); App.Server.Port == "" {
		// Setting default Port
		App.Server.Port = "80"
	}

	if App.SMB.Enabled, err = strconv.ParseBool(os.Getenv("SMB_ENABLED")); err != nil || !App.SMB.Enabled {
		App.SMB.Enabled = false
		err = nil
	} else if App.SMB.Enabled {
		if App.SMB.Servername = os.Getenv("SMB_SERVERNAME"); App.SMB.Servername == "" {
			err = errors.New("smb servername must be provided")
			return err
		}

		if App.SMB.Sharename = os.Getenv("SMB_SHARENAME"); App.SMB.Sharename == "" {
			err = errors.New("smb sharename must be provided")
			return err
		}

		if App.SMB.Username = os.Getenv("SMB_USERNAME"); App.SMB.Username == "" {
			err = errors.New("smb username must be provided")
			return err
		}

		if App.SMB.Password = os.Getenv("SMB_PASSWORD"); App.SMB.Password == "" {
			err = errors.New("smb password must be provided")
			return err
		}

		if App.SMB.Domain = os.Getenv("SMB_DOMAIN"); App.SMB.Domain == "" {
			err = errors.New("smb domain must be provided")
			return err
		}
	}

	if _, err := os.Stat("./data"); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir("./data", os.ModePerm)
		if err != nil {
			return err
		}
	}

	return err
}
