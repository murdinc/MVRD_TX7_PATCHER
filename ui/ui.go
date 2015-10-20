package ui

import (
	"fmt"
	"strings"

	ui "github.com/gizak/termui"
	"github.com/murdinc/MVRD_TX7_PATCHER/midi"
	"github.com/murdinc/MVRD_TX7_PATCHER/parse"
)

type DisplayData struct {
	// For Display Stuff
	ListIndex     int
	SelectedVoice int

	// Trigger a Sysex Send
	SendVoice bool

	// Bank and Voice Index Numbers
	BankIndex  int
	VoiceIndex int
}

func Start(l parse.Library) {
	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	// Header and Instructions
	p := ui.NewPar(" Tool for building banks of voices on the Yamaha TX7 and DX7 synths. \n\n Press 'B' and 'E' go to the beginning and end. \n\n Press 'Q' to quit.\n\n Use arrows and enter key to select and upload voice.")
	p.Height = 14
	p.Width = 75
	p.TextFgColor = ui.ColorWhite
	p.Border.Label = " MVRD_TX7_PATCHER "
	p.Border.FgColor = ui.ColorWhite
	p.Border.LabelFgColor = ui.ColorCyan

	// Information on right
	info := ui.NewPar(" ")
	info.Height = 66
	info.Width = 124
	info.TextFgColor = ui.ColorWhite
	info.Border.Label = " VOICE SETTINGS: "
	info.Border.FgColor = ui.ColorWhite
	info.Border.LabelFgColor = ui.ColorCyan
	info.Y = 0
	info.X = 76

	// List of Voices
	strs := VoiceNames(l)
	list := ui.NewList()
	list.Items = strs
	list.Border.LabelFgColor = ui.ColorCyan
	list.ItemFgColor = ui.ColorYellow
	list.Border.FgColor = ui.ColorWhite
	list.Border.Label = fmt.Sprintf(" AVAILABLE VOICES: [ %d ] in [ %d ] Banks ", l.VoiceCount, l.FileCount)
	list.Height = 52
	list.Width = 75
	list.Y = 14

	// Build output
	draw := func(displayData DisplayData) {

		bank := l.Banks[displayData.BankIndex]
		voice := bank.Voice[displayData.VoiceIndex]

		info.Text = BuildVoiceInfo(voice)

		strs = VoiceNames(l)

		voices := strs[displayData.ListIndex:displayData.SelectedVoice]
		voices = append(voices, fmt.Sprintf(">%s", strs[displayData.SelectedVoice]))
		voices = append(voices, strs[displayData.SelectedVoice+1:]...)

		list.Items = voices

		// Send Voice !
		if displayData.SendVoice == true {

			sysex := l.BuildSysex(displayData.BankIndex, displayData.VoiceIndex)
			midi.Upload(sysex)

		}

		ui.Render(p, list, info)
	}

	// Event Channel
	evt := ui.EventCh()

	listIndex := 0
	selectedVoice := 0
	sendVoice := false

	for {

		select {
		case e := <-evt:

			// Q - Quit
			if e.Type == ui.EventKey && e.Ch == 'q' {
				return
			}

			// Right Arrow - Next Page
			if e.Type == ui.EventKey && e.Key == ui.KeyArrowRight {
				selectedVoice += 49
			}

			// Left Arrow - Previous Page
			if e.Type == ui.EventKey && e.Key == ui.KeyArrowLeft {
				selectedVoice -= 49
			}

			// Down Arrow - Next Voice
			if e.Type == ui.EventKey && e.Key == ui.KeyArrowDown {
				selectedVoice++
			}

			// Up Arrow - Previous Voice
			if e.Type == ui.EventKey && e.Key == ui.KeyArrowUp {
				selectedVoice--
			}

			// E - End of list
			if e.Type == ui.EventKey && e.Ch == 'e' {
				selectedVoice = l.VoiceCount - 1
			}

			// B - Begining of list
			if e.Type == ui.EventKey && e.Ch == 'b' {
				selectedVoice = 0
			}

			// P - Previous
			if e.Type == ui.EventKey && e.Ch == 'p' {
				selectedVoice--
				sendVoice = true
			}

			// N - Next
			if e.Type == ui.EventKey && e.Ch == 'n' {
				selectedVoice++
				sendVoice = true
			}

			// Enter or Space
			if e.Type == ui.EventKey && (e.Key == ui.KeyEnter || e.Key == ui.KeySpace) {
				sendVoice = true
			}

		default:

			// Lower Limit
			if selectedVoice <= 0 {
				selectedVoice = 0
				listIndex = 0
			}

			// Upper Limit
			if selectedVoice >= l.VoiceCount {
				selectedVoice = l.VoiceCount - 1
				listIndex = l.VoiceCount - 50
			}

			// Resituate
			if selectedVoice > listIndex+49 {
				listIndex = selectedVoice - 49
			}
			if selectedVoice < listIndex {
				listIndex = selectedVoice
			}

			bankIndex := (selectedVoice * l.VoiceCount) / (l.VoiceCount) / 32
			voiceIndex := selectedVoice - (bankIndex * 32)

			displayData := DisplayData{
				ListIndex:     listIndex,
				SelectedVoice: selectedVoice,
				SendVoice:     sendVoice,
				BankIndex:     bankIndex,
				VoiceIndex:    voiceIndex,
			}

			draw(displayData)

			// Reset
			sendVoice = false

		}
	}
}

func BuildVoiceInfo(voice parse.Voice) string {

	voiceString := fmt.Sprintf(" Name: %v\n", voice.Name)

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

	voiceString += fmt.Sprintf("			Algorithm: %.2d			Feedback: %.2d			OscKeySync: %.2d\n\n", voice.Algorithm, voice.Feedback, voice.OscKeySync)

	voiceString += fmt.Sprintf("			LfoSpeed: %.2d			LfoPitchModDepth: %.2d\n", voice.LfoSpeed, voice.LfoPitchModDepth)
	voiceString += fmt.Sprintf("			LfoDelay: %.2d			LfoAMDepth: %.2d\n\n", voice.LfoDelay, voice.LfoAMDepth)

	voiceString += fmt.Sprintf("			LfoSync: %.2d\n", voice.LfoSync)
	voiceString += fmt.Sprintf("			LfoWave: %.2d\n", voice.LfoWave)
	voiceString += fmt.Sprintf("			LfoPitchModSensitivity: %.2d\n\n", voice.LfoPitchModSensitivity)

	voiceString += fmt.Sprintf("			Transpose: %.2d", voice.Transpose)

	return voiceString

}

func VoiceNames(l parse.Library) []string {

	names := make([]string, 0)

	i := 0
	for _, bank := range l.Banks {

		for _, voice := range bank.Voice {
			i++

			number := fmt.Sprintf("   # %d    ", i)
			bankName := strings.TrimPrefix(bank.FileName, l.FolderName)

			voiceName := addSpaces(number, 15) + fmt.Sprintf("%s  Bank: [%s] ", addSpaces(voice.Name, 20), bankName)

			names = append(names, voiceName)
		}

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
