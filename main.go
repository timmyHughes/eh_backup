// backup_eh project main.go
package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/urfave/cli"
)

/* -f backup_config.json
   -s awseda -k 7ff3b1a326bc491482153342ff381137 -t runningconfig  -n node1
   -s awseda -k 7ff3b1a326bc491482153342ff381137 -t customizations -n node1
*/

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type Config struct {
	Debug  *bool
	File   string
	Backup []Backup
}

type Backup struct {
	Server string `json:"server"`
	Key    string `json:"key"`
	Type   string `json:"type"`
	Node   string `json:"node"`
}

func main() {
	app := cli.NewApp()
	app.Name = "backup_eh"
	app.Version = "1.2"

	config := new(Config)
	backup := new(Backup)
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "file, f",
			Value:       "",
			Usage:       "Configuration File",
			Destination: &config.File,
		},
		cli.BoolFlag{
			Name:        "debug, d",
			Usage:       "Debug Output",
			Destination: config.Debug,
		},
		cli.StringFlag{
			Name:        "server, s",
			Usage:       "REQUIRED: Extrahop API Host",
			Destination: &backup.Server,
		},
		cli.StringFlag{
			Name:        "key, k",
			Value:       "",
			Usage:       "REQUIRED: Key for accessing Extrahop's REST API",
			Destination: &backup.Key,
		},
		cli.StringFlag{
			Name:        "type, t",
			Value:       "",
			Usage:       "REQUIRED: Backup type: [runningconfig|customizations]",
			Destination: &backup.Type,
		},
		cli.StringFlag{
			Name:        "node, n",
			Value:       "",
			Usage:       "Node Identifier",
			Destination: &backup.Node,
		},
	}

	app.Action = func(c *cli.Context) error {
		config.Backup = append(config.Backup, *backup)
		if config.File != "" {
			fileContents, e := ioutil.ReadFile(config.File)
			if e != nil {
				fmt.Printf("File error: %v\n", e)
				os.Exit(1)
			}
			json.Unmarshal(fileContents, &config)
		} else {
			if config.Backup[0].Server == "" {
				cli.ShowAppHelp(c)
				return cli.NewExitError("--server [-s] connection arguement not set\n", 99)
			}
			if config.Backup[0].Key == "" {
				cli.ShowAppHelp(c)
				return cli.NewExitError("--key [-k] connection arguement not set\n", 98)
			}
			if config.Backup[0].Type == "" {
				cli.ShowAppHelp(c)
				return cli.NewExitError("--type [-t] backup type not set\n", 97)
			}
		}

		for _, b := range config.Backup {
			loopBackups(b)
		}
		return nil
	}
	app.Run(os.Args)
}

func loopBackups(backup Backup) {
	var apiUriString string
	t := time.Now()
	file := fmt.Sprintf("./%s_%s_%d%02d%02d.json", backup.Node, backup.Type, t.Year(), t.Month(), t.Day())
	authHeaderValue := "ExtraHop apikey=" + backup.Key

	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore invalid/expired SSL certificates
	}
	client := &http.Client{Transport: transCfg}

	if backup.Type == "runningconfig" {
		apiUriString = "https://" + backup.Server + "/api/v1/" + backup.Type
	} else {
		apiUriString = "https://" + backup.Server + "/api/v1/" + backup.Type
		req, err := http.NewRequest("GET", apiUriString, nil)
		req.Header.Add("Authorization", authHeaderValue)
		rsp, err := client.Do(req)
		if err != nil {
			fmt.Println("GET Error: " + err.Error())
		} else {
			defer rsp.Body.Close()

			buf := new(bytes.Buffer)
			buf.ReadFrom(rsp.Body)
			b := buf.Bytes()

			type customization struct {
				Id          int    `json:"id"`
				Name        string `json:"name"`
				Create_Time int    `json:"create_time"`
				Auto        bool   `json:"auto"`
			}

			var customizations []customization
			err := json.Unmarshal(b, &customizations)
			check(err)
			var customization_id = ""
			for _, c := range customizations {
				customization_id = strconv.Itoa(c.Id)
			}
			apiUriString = "https://" + backup.Server + "/api/v1/" + backup.Type + "/" + customization_id
		}
	}

	f, err := os.Create(file)
	check(err)
	defer f.Close()

	req, err := http.NewRequest("GET", apiUriString, nil)
	req.Header.Add("Authorization", authHeaderValue)
	rsp, err := client.Do(req)
	if err != nil {
		fmt.Println("GET Error: " + err.Error())
	} else {
		defer rsp.Body.Close()
		_, err := io.Copy(f, rsp.Body)
		check(err)
	}
	fmt.Printf("Wrote backup to: %v\n", file)
}

func errchk(msg string, e error) {
	if e != nil {
		fmt.Printf("%s error: %v\n", msg, e)
		os.Exit(1)
	}
}
func debugPrint(location string, obj interface{}) {
	z, ex := json.MarshalIndent(obj, "", "    ")
	errchk("json.MarshalIndent", ex)
	fmt.Printf("Location: %v\n", location)
	os.Stdout.Write(z)
	fmt.Println("")
}
