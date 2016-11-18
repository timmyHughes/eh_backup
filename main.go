// backup_eh project main.go
package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	//"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/urfave/cli"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	var apiServer string
	var apiKey string
	var backupType string
	var debug *bool
	var nodeId string
	var apiUriString string

	app := cli.NewApp()
	app.Name = "backup_eh"
	app.Version = "1.0"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "server, s",
			Usage:       "REQUIRED: Extrahop API Host",
			Destination: &apiServer,
		},
		cli.StringFlag{
			Name:        "key, k",
			Value:       "",
			Usage:       "REQUIRED: Key for accessing Extrahop's REST API",
			Destination: &apiKey,
		},
		cli.StringFlag{
			Name:        "type, t",
			Value:       "",
			Usage:       "REQUIRED: Backup type: [runningconfig|customizations]",
			Destination: &backupType,
		},
		cli.BoolFlag{
			Name:        "debug, d",
			Usage:       "Debug Output",
			Destination: debug,
		},
		cli.StringFlag{
			Name:        "node, n",
			Value:       "",
			Usage:       "Node Identifier",
			Destination: &nodeId,
		},
	}

	app.Action = func(c *cli.Context) error {

		if apiServer == "" {
			cli.ShowAppHelp(c)
			return cli.NewExitError("--server [-s] connection arguement not set\n", 99)
		}
		if apiKey == "" {
			cli.ShowAppHelp(c)
			return cli.NewExitError("--key [-k] connection arguement not set\n", 98)
		}
		if backupType == "" {
			cli.ShowAppHelp(c)
			return cli.NewExitError("--type [-t] backup type not set\n", 97)
		}

		t := time.Now()
		file := fmt.Sprintf("./%s_%s_%d%02d%02d.json", backupType, nodeId, t.Year(), t.Month(), t.Day())
		authHeaderValue := "ExtraHop apikey=" + apiKey

		transCfg := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore invalid/expired SSL certificates
		}
		client := &http.Client{Transport: transCfg}

		if backupType == "runningconfig" {
			apiUriString = "https://" + apiServer + "/api/v1/" + backupType
		} else {
			apiUriString = "https://" + apiServer + "/api/v1/" + backupType
			req, err := http.NewRequest("GET", apiUriString, nil)
			req.Header.Add("Authorization", authHeaderValue)
			rsp, err := client.Do(req)
			if err != nil {
				fmt.Println("GET Error: " + err.Error())
				return nil
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
				apiUriString = "https://" + apiServer + "/api/v1/" + backupType + "/" + customization_id
			}
		}

		if c.Bool("debug") {
			fmt.Println("\nExtraHop API URI: " + apiUriString)
			fmt.Println("Authorization Value: " + authHeaderValue + "\n")
			fmt.Println("Output File: " + file + "\n")
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
		return nil
	}
	app.Run(os.Args)
}
