package main

import (
	"fmt"
	"os"

	"./parse"
	"github.com/murdinc/cli"
)

// Main Function
////////////////..........
func main() {

	app := cli.NewApp()
	app.Name = "TX7 Patcher"
	app.Usage = "Command Line Interface for managing patches on the Yamaha DX7 and TX7 Synthesizers"
	app.Version = "1.0"
	app.Commands = []cli.Command{
		{
			Name:        "download",
			ShortName:   "d",
			Example:     "download",
			Description: "Download the calibration from the ECU",
			Action: func(c *cli.Context) {
				//obd := iso9141.New()
				//obd.DownloadBIN("DOWNLOAD")
			},
		},
		{
			Name:        "parse",
			ShortName:   "p",
			Example:     "parse patch.syx",
			Description: "Parse a sysex file and display contents",
			Arguments: []cli.Argument{
				cli.Argument{Name: "sysex", Usage: "parse patch.syx", Description: "The name of the sysex file to parse", Optional: false},
			},
			Action: func(c *cli.Context) {
				parse, _ := parse.New(c.NamedArg("sysex"))
				parse.DisplayPatches()
			},
		},
	}

	log("TX7 Patcher - v1.0", nil)
	log("Created by Ahmad A.", nil)
	log("© MVRD INDUSTRIES 2015", nil)
	log("Not for commercial use", nil)
	app.Run(os.Args)
}

// Log Function
////////////////..........
func log(kind string, err error) {
	if err == nil {
		fmt.Printf("====> %s\n", kind)
	} else {
		fmt.Printf("[ERROR - %s]: %s\n", kind, err)
	}
}
