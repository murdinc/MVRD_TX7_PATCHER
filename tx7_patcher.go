package main

import (
	"fmt"
	"os"

	"github.com/murdinc/MVRD_TX7_PATCHER/parse"
	"github.com/murdinc/MVRD_TX7_PATCHER/tx7"
	"github.com/murdinc/MVRD_TX7_PATCHER/ui"
	"github.com/murdinc/cli"
	"github.com/murdinc/terminal"
)

// Main Function
////////////////..........
func main() {

	app := cli.NewApp()
	app.Name = "TX7 Patcher"
	app.Version = "1.0"
	app.Commands = []cli.Command{
		{
			Name:        "parse",
			ShortName:   "p",
			Description: "Parse a sysex file and display contents",
			Arguments: []cli.Argument{
				{Name: "sysex", Usage: "parse patch.syx", Description: "The name of the sysex file to parse", Optional: false},
			},
			Action: func(c *cli.Context) error {
				bank, _, _ := parse.Open(c.NamedArg("sysex"), &map[uint64]string{})
				bank.DisplayVoices()
				return nil
			},
		},
		{
			Name:        "test",
			ShortName:   "t",
			Description: "Parse all sysex files in a directory and test program",
			Arguments: []cli.Argument{
				{Name: "folder", Usage: "test /foldername", Description: "The name of the sysex folder to test against", Optional: false},
			},
			Action: func(c *cli.Context) error {
				library, _ := parse.OpenDir(c.NamedArg("folder"))
				/*
					for _, bank := range library.Banks {

						terminal.Information("		" + bank.FileName)

						for _, voice := range bank.Voice {
							terminal.Information(voice.Name)
						}

					}
				*/
				terminal.Information(fmt.Sprintf("Duplicates: %d", library.Duplicates))

				return nil
			},
		},
		{
			Name:        "run",
			ShortName:   "r",
			Description: "Parse all sysex files in a directory and start program",
			Arguments: []cli.Argument{
				{Name: "folder", Usage: "run /foldername", Description: "The name of the sysex folder to run against", Optional: false},
			},
			Action: func(c *cli.Context) error {
				library, _ := parse.OpenDir(c.NamedArg("folder"))

				// Get device id's
				input, output, err := tx7.Discover()
				if err != nil {
					return err
				}

				synth, err := tx7.New(input, output)
				if err != nil {
					return err
				}

				synth.Open()

				ui.Start(library, synth)
				return nil
			},
		},
		{
			Name:        "listVoiceNames",
			ShortName:   "lvn",
			Description: "List all voice names of all the sysex files in a directory",
			Arguments: []cli.Argument{
				{Name: "folder", Usage: "listVoiceNames /foldername", Description: "The name of the sysex folder to parse", Optional: false},
			},
			Action: func(c *cli.Context) error {
				library, _ := parse.OpenDir(c.NamedArg("folder"))
				library.DisplayVoiceNames()
				return nil
			},
		},
		{
			Name:        "upload",
			ShortName:   "u",
			Description: "upload",
			Arguments: []cli.Argument{
				{Name: "sysex", Usage: "upload ./sysex/WEIRD1.SYX", Description: "The name of the sysex bank file to upload", Optional: false},
			},
			Action: func(c *cli.Context) error {
				sysex, _, _ := parse.Open(c.NamedArg("sysex"), &map[uint64]string{})

				// Get device id's
				input, output, err := tx7.Discover()
				if err != nil {
					return err
				}

				synth, err := tx7.New(input, output)
				if err != nil {
					return err
				}

				synth.Open()

				synth.Upload(sysex.Raw)

				return nil
			},
		},
		{
			Name:        "displayVoice",
			ShortName:   "dv",
			Description: "Download the currently selected voice and Display it",
			Action: func(c *cli.Context) error {

				callback := func(sysexBytes []byte) {
					bank, _ := parse.New(sysexBytes)
					bank.DisplayVoices()
				}

				// Get device id's
				input, output, err := tx7.Discover()
				if err != nil {
					return err
				}

				synth, err := tx7.New(input, output)
				if err != nil {
					return err
				}

				synth.Open()

				synth.DownloadVoice(callback)

				return nil
			},
		},
		{
			Name:        "displayBank",
			ShortName:   "db",
			Description: "Download the bank and Display it",
			Action: func(c *cli.Context) error {

				callback := func(sysexBytes []byte) {
					bank, _ := parse.New(sysexBytes)
					bank.DisplayVoices()
				}

				// Get device id's
				input, output, err := tx7.Discover()
				if err != nil {
					return err
				}

				synth, err := tx7.New(input, output)
				if err != nil {
					return err
				}

				synth.Open()

				synth.DownloadBank(callback)

				return nil
			},
		},
	}

	log("TX7 Patcher - v1.2", nil)
	log("Created by Ahmad A.", nil)
	log("© MVRD INDUSTRIES 2015, 2016, 2017", nil)
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
