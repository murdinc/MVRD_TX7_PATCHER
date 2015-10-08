package main

import (
	"fmt"
	"os"

	"github.com/murdinc/MVRD_TX7_PATCHER/midi"
	"github.com/murdinc/MVRD_TX7_PATCHER/parse"
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
		{
			Name:        "parseDirectory",
			ShortName:   "pd",
			Example:     "parse /foldername",
			Description: "Parse all sysex files in a directory and display contents",
			Arguments: []cli.Argument{
				cli.Argument{Name: "folder", Usage: "parse /foldername", Description: "The name of the sysex folder to parse", Optional: false},
			},
			Action: func(c *cli.Context) {
				parse, _ := parse.New(c.NamedArg("folder"))
				parse.DisplayPatches()
			},
		},
		{
			Name:        "monitor",
			ShortName:   "m",
			Example:     "monitor",
			Description: "Monitors all midi in messages",
			Action: func(c *cli.Context) {
				midi, _ := midi.New()
				midi.Monitor()
			},
		},
		{
			Name:        "itentity",
			ShortName:   "i",
			Example:     "itentity",
			Description: "Lists Identities of connected devices",
			Action: func(c *cli.Context) {
				midi, _ := midi.New()
				midi.Identity()
			},
		},
		{
			Name:        "destinations",
			ShortName:   "d",
			Example:     "destinations",
			Description: "List all destinations",
			Action: func(c *cli.Context) {
				midi, _ := midi.New()
				midi.ListDestinations()
			},
		},
		{
			Name:        "sources",
			ShortName:   "s",
			Example:     "sources",
			Description: "List all Sources",
			Action: func(c *cli.Context) {
				midi, _ := midi.New()
				midi.ListSources()
			},
		},
		{
			Name:        "test",
			ShortName:   "t",
			Example:     "test",
			Description: "test send",
			Arguments: []cli.Argument{
				cli.Argument{Name: "sysex", Usage: "parse patch.syx", Description: "The name of the sysex file to send", Optional: false},
			},
			Action: func(c *cli.Context) {
				sysex, _ := parse.New(c.NamedArg("sysex"))
				midi, _ := midi.New()
				midi.TestSend(sysex.Raw)
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
