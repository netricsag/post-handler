package main

import (
	"io/ioutil"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/basicauth"

	"github.com/hirochachacha/go-smb2"
	"github.com/natron-io/post-handler/util"
)

func Init() {
	util.InitLoggers()

	if err := util.LoadEnv(); err != nil {
		util.ErrorLogger.Fatal(err)
	}

	util.InfoLogger.Println("Config loaded")
}

func main() {

	srv := fiber.New()

	srv.Use(basicauth.New(basicauth.Config{
		Users: map[string]string{
			util.App.Auth.Username: util.App.Auth.Password,
		},
	}))

	srv.Post("/upload", getDataStream)

	if err := srv.Listen(":8080"); err != nil {
		util.ErrorLogger.Fatal(err)
	}
}

func getDataStream(c *fiber.Ctx) error {
	// Get the data bytes
	bodyBytes := c.Body()
	if len(bodyBytes) == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "No data received",
		})
	}

	nameOfFile, err := writeToFile(bodyBytes)
	if err != nil {
		return err
	}
	util.InfoLogger.Printf("Local Storage: Sucessfully uploaded to %s", nameOfFile)

	if util.App.SMB.Enabled {
		err := pushOnSMB(nameOfFile, bodyBytes)
		if err != nil {
			return err
		}
		util.InfoLogger.Printf("SMB: Sucessfully uploaded to \\\\%s\\%s\\%s", util.App.SMB.Servername, util.App.SMB.Sharename, nameOfFile)
	}

	return c.Status(200).JSON(fiber.Map{
		"success": true,
	})
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

func pushOnSMB(filename string, fileContent []byte) error {
	connection, err := net.Dial("tcp", util.App.SMB.Servername+":445")
	if err != nil {
		return err
	}

	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     util.App.SMB.Username,
			Password: util.App.SMB.Password,
			Domain:   util.App.SMB.Domain,
		},
	}

	sambaClient, err := d.Dial(connection)
	if err != nil {
		return err
	}
	defer sambaClient.Logoff()

	fs, err := sambaClient.Mount("\\\\" + util.App.SMB.Servername + "\\" + util.App.SMB.Sharename)
	if err != nil {
		return err
	}
	defer fs.Umount()

	if err = fs.WriteFile(filename, fileContent, 0444); err != nil {
		return err
	}

	return nil
}
