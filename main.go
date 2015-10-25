package main

import (
	"fmt"
	"os"

	"github.com/murdinc/MVRD_TX7_PATCHER/midi"
	"github.com/murdinc/MVRD_TX7_PATCHER/parse"
	"github.com/murdinc/MVRD_TX7_PATCHER/ui"
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
				bank, _ := parse.Open(c.NamedArg("sysex"))
				bank.DisplayVoices()
			},
		},
		{
			Name:        "run",
			ShortName:   "r",
			Example:     "run /foldername",
			Description: "Parse all sysex files in a directory and start program",
			Arguments: []cli.Argument{
				cli.Argument{Name: "folder", Usage: "run /foldername", Description: "The name of the sysex folder to run against", Optional: false},
			},
			Action: func(c *cli.Context) {
				library, _ := parse.OpenDir(c.NamedArg("folder"))
				ui.Start(library)
			},
		},
		{
			Name:        "listFiles",
			ShortName:   "lf",
			Example:     "listFiles /foldername",
			Description: "List all sysex files in a directory and display contents",
			Arguments: []cli.Argument{
				cli.Argument{Name: "folder", Usage: "listFiles /foldername", Description: "The name of the sysex folder to parse", Optional: false},
			},
			Action: func(c *cli.Context) {
				library, _ := parse.OpenDir(c.NamedArg("folder"))
				library.DisplayFileNames()
			},
		},
		{
			Name:        "listVoiceNames",
			ShortName:   "lvn",
			Example:     "listVoiceNames /foldername",
			Description: "List all voice names of all the sysex files in a directory",
			Arguments: []cli.Argument{
				cli.Argument{Name: "folder", Usage: "listVoiceNames /foldername", Description: "The name of the sysex folder to parse", Optional: false},
			},
			Action: func(c *cli.Context) {
				library, _ := parse.OpenDir(c.NamedArg("folder"))
				library.DisplayVoiceNames()
			},
		},
		{
			Name:        "upload",
			ShortName:   "u",
			Example:     "upload ./sysex/WEIRD1.SYX",
			Description: "upload",
			Arguments: []cli.Argument{
				cli.Argument{Name: "sysex", Usage: "upload ./sysex/WEIRD1.SYX", Description: "The name of the sysex bank file to upload", Optional: false},
			},
			Action: func(c *cli.Context) {
				sysex, _ := parse.Open(c.NamedArg("sysex"))
				midi.Upload(sysex.Raw)
			},
		},
		{
			Name:        "displayVoice",
			ShortName:   "dv",
			Example:     "displayVoice",
			Description: "Download the currently selected voice and Display it",
			Action: func(c *cli.Context) {

				callback := func(sysexBytes []byte) {
					bank, _ := parse.New(sysexBytes)
					bank.DisplayVoices()
				}

				midi.DownloadVoice(callback)
			},
		},
		{
			Name:        "displayBank",
			ShortName:   "db",
			Example:     "displayBank",
			Description: "Download the bank and Display it",
			Action: func(c *cli.Context) {

				callback := func(sysexBytes []byte) {
					bank, _ := parse.New(sysexBytes)
					bank.DisplayVoices()
				}

				midi.DownloadBank(callback)
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
