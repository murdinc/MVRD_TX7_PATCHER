package ui

import (
	"fmt"
	"strings"

	ui "github.com/gizak/termui"
	"github.com/murdinc/MVRD_TX7_PATCHER/parse"
	"github.com/murdinc/MVRD_TX7_PATCHER/tx7"
)

func Start(l parse.Library, synth *tx7.TX7) {
	if err := ui.Init(); err != nil {
		panic(err)
	}
	defer ui.Close()
	defer synth.Close()

	voiceList := l.Voices()
	voiceCount := l.VoiceCount()

	// Header and Instructions
	header := ui.NewPar(" ")
	header.Height = 14
	header.Width = 75
	header.TextFgColor = ui.ColorWhite
	header.BorderLabel = " MVRD_TX7_PATCHER "
	header.BorderFg = ui.ColorWhite
	header.BorderLabelFg = ui.ColorCyan

	// List of Voices
	list := ui.NewList()
	list.BorderLabelFg = ui.ColorCyan
	list.ItemFgColor = ui.ColorYellow
	list.BorderFg = ui.ColorWhite
	list.Height = 52
	list.Width = 75
	list.Y = 14

	// Information on right
	info := ui.NewPar(" ")
	info.Height = 66
	info.Width = 124
	info.TextFgColor = ui.ColorWhite
	info.BorderLabel = " VOICE SETTINGS: "
	info.BorderFg = ui.ColorWhite
	info.BorderLabelFg = ui.ColorCyan
	info.Y = 0
	info.X = 76

	// Scroll Bar
	scroll := ui.NewGauge()
	scroll.Percent = 0
	scroll.Width = 200
	scroll.Height = 3
	scroll.BarColor = ui.ColorRed
	scroll.BorderFg = ui.ColorWhite
	scroll.Y = 66

	listIndex := 0
	selectedVoice := 0
	search := false
	searchStr := ""

	// Build output
	draw := func(listIndex int, selectedVoice int, sendVoice bool, search bool, searchStr string) {

		header.Text = " Tool for building banks of voices on the Yamaha TX7 and DX7 synths. \n\n Press 'B' and 'E' go to the beginning and end. \n\n Press 'Q' to quit.\n\n Use arrows and enter key to select and upload voice."
		header.Text += "\n\n	Search [ " + searchStr + " ]"
		if search == true {
			header.Text += " < \n\n	Search Mode! Press ESC to exit."
		} else {
			header.Text += fmt.Sprintf("\n\n Selected Voice: %d", selectedVoice+1)
		}

		if l.SearchStr != searchStr {
			l.Search(searchStr)

			voiceList = l.Voices()
			voiceCount = l.VoiceCount()
		}

		if voiceCount > 0 {

			// Resituate
			if selectedVoice > listIndex+49 {
				listIndex = selectedVoice - 49
			}
			if selectedVoice < listIndex {
				listIndex = selectedVoice
			}

			strs := VoiceNames(l)

			if selectedVoice >= 0 && selectedVoice < len(voiceList) && selectedVoice < len(strs) {
				info.Text = BuildVoiceInfo(voiceList[selectedVoice])

				voices := strs[listIndex:selectedVoice]
				voices = append(voices, fmt.Sprintf(">%s", strs[selectedVoice]))
				voices = append(voices, strs[selectedVoice+1:]...)
				list.Items = voices
				scroll.Percent = ((selectedVoice + 1) * 100) / voiceCount
			}

			// Send Voice !
			if sendVoice == true {
				sysex := l.BuildSysex(selectedVoice)
				synth.Upload(sysex)
			}

		} else {
			list.Items = []string{"		No Items Found !		"}
		}

		list.BorderLabel = fmt.Sprintf(" VOICES: [ %d ] in [ %d ] Banks ( [%d] Duplicate Voices ) ", voiceCount, l.FileCount, l.Duplicates)

		ui.Render(header, list, info, scroll)
	}

	draw(listIndex, selectedVoice, false, search, searchStr)

	// S - Search
	ui.Handle("/sys/kbd/s", func(ui.Event) {
		if search != true {
			search = true
		} else {
			searchStr += "s"
		}

		draw(listIndex, selectedVoice, false, search, searchStr)
	})

	// ESC - Escape Search
	ui.Handle("/sys/kbd/<escape>", func(ui.Event) {
		if search == true {
			search = false
			draw(listIndex, selectedVoice, false, search, searchStr)
		}
	})

	// Delete - Delete in Search
	ui.Handle("/sys/kbd/C-8", func(ui.Event) {
		if search == true {

			sz := len(searchStr)
			if sz > 0 {
				searchStr = searchStr[:sz-1]
			}

			draw(listIndex, selectedVoice, false, search, searchStr)
		}
	})

	// Keys
	ui.Handle("/sys/kbd/", func(e ui.Event) {
		if search == true {
			//fmt.Println(e)

			key := strings.TrimPrefix(e.Path, "/sys/kbd/")
			if len(key) == 1 {
				searchStr += key
				selectedVoice = 0
				listIndex = 0
			}
		}

		draw(listIndex, selectedVoice, false, search, searchStr)
	})

	// handle key q pressing
	ui.Handle("/sys/kbd/q", func(ui.Event) {
		if search != true {
			// press q to quit
			ui.StopLoop()
		} else {
			searchStr += "q"
		}
	})

	// Right Arrow - Next Page
	ui.Handle("/sys/kbd/<right>", func(ui.Event) {
		selectedVoice += 49

		// Upper Limit
		if selectedVoice >= voiceCount {
			selectedVoice = voiceCount - 1
			if l.VoiceCount() > 50 {
				listIndex = voiceCount - 50
			} else {
				listIndex = 0
			}
		}

		draw(listIndex, selectedVoice, false, search, searchStr)
	})

	// Left Arrow - Previous Page
	ui.Handle("/sys/kbd/<left>", func(ui.Event) {
		selectedVoice -= 49

		// Lower Limit
		if selectedVoice <= 0 {
			selectedVoice = 0
			listIndex = 0
		}

		draw(listIndex, selectedVoice, false, search, searchStr)
	})

	// Down Arrow - Next Voice
	ui.Handle("/sys/kbd/<down>", func(ui.Event) {
		if selectedVoice < voiceCount-1 {
			selectedVoice++
		} else {
			selectedVoice = voiceCount - 1
		}
		draw(listIndex, selectedVoice, false, search, searchStr)
	})

	// Up Arrow - Previous Voice
	ui.Handle("/sys/kbd/<up>", func(ui.Event) {
		if selectedVoice > 0 {
			selectedVoice--
		} else {
			selectedVoice = 0
		}
		draw(listIndex, selectedVoice, false, search, searchStr)
	})

	// E - End of list
	ui.Handle("/sys/kbd/e", func(ui.Event) {
		if search == false {
			selectedVoice = l.VoiceCount() - 1
		} else {
			searchStr += "e"
		}
		draw(listIndex, selectedVoice, false, search, searchStr)

	})

	// B - Begining of list
	ui.Handle("/sys/kbd/b", func(ui.Event) {
		if search == false {
			selectedVoice = 0
		} else {
			searchStr += "b"
		}
		draw(listIndex, selectedVoice, false, search, searchStr)

	})

	// P - Previous
	ui.Handle("/sys/kbd/p", func(ui.Event) {
		if search == false {
			if selectedVoice > 0 {
				selectedVoice--
			} else {
				selectedVoice = 0
			}
			draw(listIndex, selectedVoice, true, search, searchStr)
		} else {
			searchStr += "p"
			draw(listIndex, selectedVoice, false, search, searchStr)

		}
	})

	// N - Next
	ui.Handle("/sys/kbd/n", func(ui.Event) {
		if search == false {
			if selectedVoice < voiceCount-1 {
				selectedVoice++
			} else {
				selectedVoice = voiceCount - 1
			}
			draw(listIndex, selectedVoice, true, search, searchStr)
		} else {
			searchStr += "n"
			draw(listIndex, selectedVoice, false, search, searchStr)
		}
	})

	// Enter
	ui.Handle("/sys/kbd/<enter>", func(ui.Event) {
		if search == false {
			draw(listIndex, selectedVoice, true, search, searchStr)
		}

	})

	// Space
	ui.Handle("/sys/kbd/<space>", func(ui.Event) {
		if search == false {
			draw(listIndex, selectedVoice, true, search, searchStr)
		} else {
			searchStr += " "
			draw(listIndex, selectedVoice, false, search, searchStr)
		}
	})

	ui.Loop()

}

func BuildVoiceInfo(voice parse.Voice) string {

	voiceString := fmt.Sprintf(" Name: %v\n", voice.Name)
	voiceString += fmt.Sprintf(" Filename: %v\n", voice.BankFileName)

	// Start Operators
	for n, operator := range voice.Operators {

		voiceString += fmt.Sprintf("		Operator %d\n", n+1)

		voiceString += fmt.Sprintf("				EGRate1: %.3d		EGLevel1: %.3d		ScaleLeftDepth:  %.3d", operator.EGRate1, operator.EGLevel1, operator.ScaleLeftDepth)
		voiceString += fmt.Sprintf("				LevelScalingBreakPoint: %.2d			AmplitudeModulationSensitivity: %d\n", operator.LevelScalingBreakPoint, operator.AmplitudeModulationSensitivity)

		voiceString += fmt.Sprintf("				EGRate2: %.3d		EGLevel2: %.3d		ScaleRightDepth: %.3d", operator.EGRate2, operator.EGLevel2, operator.ScaleRightDepth)
		voiceString += fmt.Sprintf("				RateScale: %d																	KeyVelocitySensitivity: %d\n", operator.RateScale, operator.KeyVelocitySensitivity)

		voiceString += fmt.Sprintf("				EGRate3: %.3d		EGLevel3: %.3d		ScaleLeftCurve:  %.3d", operator.EGRate3, operator.EGLevel3, operator.ScaleLeftCurve)
		voiceString += fmt.Sprintf("				RateScale: %d																	KeyVelocitySensitivity: %d\n", operator.RateScale, operator.KeyVelocitySensitivity)

		voiceString += fmt.Sprintf("				EGRate4: %.3d		EGLevel4: %.3d		ScaleRightCurve: %.3d", operator.EGRate4, operator.EGLevel4, operator.ScaleRightCurve)
		voiceString += fmt.Sprintf("				Detune: %.2d																			OutputLevel: %d\n", operator.Detune, operator.OutputLevel)

		voiceString += fmt.Sprintf("                                                         FrequencyCoarse: %.2d          OscillatorMode: %d\n", operator.FrequencyCoarse, operator.OscillatorMode)
		voiceString += fmt.Sprintf("                                                         FrequencyFine: %.2d\n\n", operator.FrequencyFine)

	}
	// End Operators

	voiceString += fmt.Sprintf("			PitchEGRate1: %.2d		PitchEGLevel1: %.2d\n", voice.PitchEGRate1, voice.PitchEGLevel1)

	voiceString += fmt.Sprintf("			PitchEGRate2: %.2d		PitchEGLevel2: %.2d\n", voice.PitchEGRate2, voice.PitchEGLevel2)
	voiceString += fmt.Sprintf("			PitchEGRate3: %.2d		PitchEGLevel3: %.2d\n", voice.PitchEGRate3, voice.PitchEGLevel3)
	voiceString += fmt.Sprintf("			PitchEGRate4: %.2d		PitchEGLevel4: %.2d\n\n", voice.PitchEGRate4, voice.PitchEGLevel4)

	voiceString += fmt.Sprintf("			Algorithm: %.2d			Feedback: %.2d			OscKeySync: %.2d			Transpose: %.2d\n\n", voice.Algorithm, voice.Feedback, voice.OscKeySync, voice.Transpose)

	voiceString += fmt.Sprintf("			LfoSpeed: %.2d			LfoPitchModDepth: %.2d\n", voice.LfoSpeed, voice.LfoPitchModDepth)
	voiceString += fmt.Sprintf("			LfoDelay: %.2d			LfoAMDepth: %.2d\n\n", voice.LfoDelay, voice.LfoAMDepth)

	voiceString += fmt.Sprintf("			LfoSync: %.2d\n", voice.LfoSync)
	voiceString += fmt.Sprintf("			LfoWave: %.2d\n", voice.LfoWave)
	voiceString += fmt.Sprintf("			LfoPitchModSensitivity: %.2d\n\n", voice.LfoPitchModSensitivity)

	return voiceString

}

func VoiceNames(l parse.Library) []string {

	names := make([]string, 0)

	i := 0

	for _, voice := range l.Voices() {
		i++

		number := fmt.Sprintf("   # %d    ", i)
		bankName := strings.TrimPrefix(voice.BankFileName, l.FolderName)

		voiceName := addSpaces(number, 15) + fmt.Sprintf("%s  Bank: [%s] ", addSpaces(voice.Name, 20), bankName)

		names = append(names, voiceName)
	}

	return names

}

func addSpaces(s string, w int) string {
	if len(s) < w {
		s += strings.Repeat(" ", w-len(s))
	}
	return s
}

func addSpacesL(s string, w int) string {
	l := ""
	if len(s) < w {
		l += strings.Repeat(" ", w-len(s))
	}
	l += s
	return l
}

func log(kind string, err error) {
	if err == nil {
		fmt.Printf("   %s\n", kind)
	} else {
		fmt.Printf("[ERROR - %s]: %s\n", kind, err)
	}
}
