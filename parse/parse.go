package parse

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Library struct {
	Banks      []Bank
	FileCount  int
	VoiceCount int
	FolderName string
}

type Bank struct {
	FileName         string
	VoiceCount       int
	Voice            []Voice
	Raw              []byte
	Start            byte
	Manufacturer     byte
	StatusAndChannel byte
	Format           byte
	Size             int16
	Checksum         byte
	End              byte
}

type Voice struct {
	Operators              []Operator
	PitchEGRate1           byte
	PitchEGRate2           byte
	PitchEGRate3           byte
	PitchEGRate4           byte
	PitchEGLevel1          byte
	PitchEGLevel2          byte
	PitchEGLevel3          byte
	PitchEGLevel4          byte
	Algorithm              byte
	Feedback               byte
	OscKeySync             byte
	LfoSpeed               byte
	LfoDelay               byte
	LfoPitchModDepth       byte
	LfoAMDepth             byte
	LfoSync                byte
	LfoWave                byte
	LfoPitchModSensitivity byte
	Transpose              byte
	Name                   string
}

type Operator struct {
	EGRate1                        byte
	EGRate2                        byte
	EGRate3                        byte
	EGRate4                        byte
	EGLevel1                       byte
	EGLevel2                       byte
	EGLevel3                       byte
	EGLevel4                       byte
	LevelScalingBreakPoint         byte
	ScaleLeftDepth                 byte
	ScaleRightDepth                byte
	ScaleLeftCurve                 byte
	ScaleRightCurve                byte
	RateScale                      byte
	Detune                         byte
	AmplitudeModulationSensitivity byte
	KeyVelocitySensitivity         byte
	OutputLevel                    byte
	OscillatorMode                 byte
	FrequencyCoarse                byte
	FrequencyFine                  byte
}

// App constants
////////////////..........
const debug = false

func OpenDir(foldername string) (Library, error) {

	files, err := ioutil.ReadDir(foldername)
	if err != nil {
		log("OpenDir", err)
	}

	banks := make([]Bank, 0)
	voiceCount := 0

	for _, file := range files {
		if strings.HasSuffix(strings.ToLower(file.Name()), ".syx") {
			//log(fmt.Sprintf("Scanning File: [%s]", file.Name()), nil)

			bank, err := Open(foldername + file.Name())

			if err == nil {
				banks = append(banks, bank)
				voiceCount += 32
				bank.Parse()
			}
		}
	}

	library := Library{Banks: banks, FileCount: len(banks), VoiceCount: voiceCount, FolderName: foldername}

	log(fmt.Sprintf("Files:  [ %d ]", library.FileCount), nil)

	log(fmt.Sprintf("Voices: [ %d ]", library.VoiceCount), nil)

	return library, nil

}

func (l *Library) Length() int {

	return len(l.Banks)

}

func (l *Library) DisplayFileNames() {

	i := 0
	for _, bank := range l.Banks {
		i++
		log(fmt.Sprintf("[# %d]		File Name: [%s]", i, bank.FileName), nil)
	}

}

func (l *Library) DisplayVoiceNames() {

	i := 0
	for _, bank := range l.Banks {

		for _, voice := range bank.Voice {
			i++
			log(fmt.Sprintf("[# %d]		Voice Name: [%s]		File Name: [%s] ", i, voice.Name, bank.FileName), nil)
		}

	}
}

func Open(fileName string) (Bank, error) {

	f, err := os.Open(fileName)
	fi, err := f.Stat()
	if err != nil {
		log(fmt.Sprintf("Open - Error opening bank file %s", fileName), err)
		return Bank{}, err
	}
	defer f.Close()

	fileSize := fi.Size()
	sysexFile := make([]byte, fileSize)

	// Read in all the bytes
	_, err = f.Read(sysexFile)
	if err != nil {
		log("Open - Error reading bank file", err)
		return Bank{}, err
	}
	//log(fmt.Sprintf("Open - reading %d bytes from bank file.", n), nil)
	//log(fmt.Sprintf("Open - [%s] is %d bytes long", fileName, fileSize), nil)

	bank := Bank{Raw: sysexFile, FileName: fileName}

	err = bank.Parse()
	if err != nil {
		log("Open - Error parsing bank file", err)
		return Bank{}, err
	}

	return bank, nil
}

func New(raw []byte) (Bank, error) {
	bank := Bank{Raw: raw}

	err := bank.Parse()
	if err != nil {
		log("New - Error parsing bank file", err)
		return Bank{}, err
	}

	return bank, nil
}

func (bank *Bank) Parse() error {
	bank.Start = bank.Raw[0] //F0
	bank.Manufacturer = bank.Raw[1]
	bank.StatusAndChannel = bank.Raw[2]
	bank.Format = bank.Raw[3]
	bank.Size = int16((int16(bank.Raw[4]) << 7) | int16(bank.Raw[5]))

	switch bank.Format {
	case 0x00:
		bank.VoiceCount = 1
		bank.doSingleVoice()
	default:
		bank.VoiceCount = 32
		bank.doBulkVoices()
	}

	bank.Checksum = bank.Raw[bank.Size+6]
	bank.End = bank.Raw[bank.Size+7] //F7

	return nil
}

func (bank *Bank) doSingleVoice() {

	voiceStart := 6

	bank.Voice = make([]Voice, 1)

	bank.Voice[0] = Voice{

		Operators: doSingleVoiceOperators(bank.Raw[voiceStart : voiceStart+126]),

		PitchEGRate1:  bank.Raw[voiceStart+126],
		PitchEGRate2:  bank.Raw[voiceStart+127],
		PitchEGRate3:  bank.Raw[voiceStart+128],
		PitchEGRate4:  bank.Raw[voiceStart+129],
		PitchEGLevel1: bank.Raw[voiceStart+130],
		PitchEGLevel2: bank.Raw[voiceStart+131],
		PitchEGLevel3: bank.Raw[voiceStart+132],
		PitchEGLevel4: bank.Raw[voiceStart+133],

		Algorithm: bank.Raw[voiceStart+134],

		Feedback:   bank.Raw[voiceStart+135],
		OscKeySync: bank.Raw[voiceStart+136],

		LfoSpeed: bank.Raw[voiceStart+137],
		LfoDelay: bank.Raw[voiceStart+138],

		LfoPitchModDepth: bank.Raw[voiceStart+139],
		LfoAMDepth:       bank.Raw[voiceStart+140],

		LfoSync:                bank.Raw[voiceStart+141],
		LfoWave:                bank.Raw[voiceStart+142],
		LfoPitchModSensitivity: bank.Raw[voiceStart+143],

		Transpose: bank.Raw[voiceStart+144],

		Name: string(bank.Raw[voiceStart+145 : voiceStart+155]),
	}

}

func (bank *Bank) doBulkVoices() {
	bank.Voice = make([]Voice, 32)

	voiceStart := 6

	for i := 0; i < bank.VoiceCount; i++ {

		bank.Voice[i] = Voice{

			Operators: doBulkOperators(bank.Raw[voiceStart : voiceStart+102]),

			PitchEGRate1:  bank.Raw[voiceStart+102],
			PitchEGRate2:  bank.Raw[voiceStart+103],
			PitchEGRate3:  bank.Raw[voiceStart+104],
			PitchEGRate4:  bank.Raw[voiceStart+105],
			PitchEGLevel1: bank.Raw[voiceStart+106],
			PitchEGLevel2: bank.Raw[voiceStart+107],
			PitchEGLevel3: bank.Raw[voiceStart+108],
			PitchEGLevel4: bank.Raw[voiceStart+109],

			Algorithm: bank.Raw[voiceStart+110],

			Feedback:   bank.Raw[voiceStart+111] & 0x7,        // bits 0 - 2
			OscKeySync: (bank.Raw[voiceStart+111] & 0x8) >> 3, // bit 3

			LfoSpeed: bank.Raw[voiceStart+112],
			LfoDelay: bank.Raw[voiceStart+113],

			LfoPitchModDepth: bank.Raw[voiceStart+114],
			LfoAMDepth:       bank.Raw[voiceStart+115],

			LfoSync:                bank.Raw[voiceStart+116] & 0x1,         // bit 0
			LfoWave:                (bank.Raw[voiceStart+116] & 0x1E) >> 1, // bits 1 - 4
			LfoPitchModSensitivity: (bank.Raw[voiceStart+116] & 0x60) >> 5, // bits 5 - 6

			Transpose: bank.Raw[voiceStart+117],

			Name: string(bank.Raw[voiceStart+118 : voiceStart+128]),
		}

		voiceStart += 128
	}

}

func doBulkOperators(raw []byte) []Operator {
	operators := make([]Operator, 6)

	operatorStart := 0

	for i := 0; i < 6; i++ {
		operators[i] = Operator{
			EGRate1:                raw[operatorStart],
			EGRate2:                raw[operatorStart+1],
			EGRate3:                raw[operatorStart+2],
			EGRate4:                raw[operatorStart+3],
			EGLevel1:               raw[operatorStart+4],
			EGLevel2:               raw[operatorStart+5],
			EGLevel3:               raw[operatorStart+6],
			EGLevel4:               raw[operatorStart+7],
			LevelScalingBreakPoint: raw[operatorStart+8],
			ScaleLeftDepth:         raw[operatorStart+9],
			ScaleRightDepth:        raw[operatorStart+10],

			ScaleLeftCurve:  raw[operatorStart+11] & 0x3,        // bits 0 - 1
			ScaleRightCurve: (raw[operatorStart+11] & 0xC) >> 2, // bits 2 - 3

			RateScale: raw[operatorStart+12] & 0x7,         // bits 0 - 2
			Detune:    (raw[operatorStart+12] & 0x78) >> 3, // bits 3 - 6

			AmplitudeModulationSensitivity: raw[operatorStart+13] & 0x3,         // bits 0 - 1
			KeyVelocitySensitivity:         (raw[operatorStart+13] & 0x1C) >> 2, // bites 2 - 4

			OutputLevel: raw[operatorStart+14],

			OscillatorMode:  raw[operatorStart+15] & 0x1,         // bit 0
			FrequencyCoarse: (raw[operatorStart+15] & 0x3E) >> 1, // bits 1 - 5

			FrequencyFine: raw[operatorStart+16],
		}
		operatorStart += 17
	}

	return operators
}

func doSingleVoiceOperators(raw []byte) []Operator {
	operators := make([]Operator, 6)

	operatorStart := 0

	for i := 0; i < 6; i++ {
		operators[i] = Operator{
			EGRate1:                raw[operatorStart],
			EGRate2:                raw[operatorStart+1],
			EGRate3:                raw[operatorStart+2],
			EGRate4:                raw[operatorStart+3],
			EGLevel1:               raw[operatorStart+4],
			EGLevel2:               raw[operatorStart+5],
			EGLevel3:               raw[operatorStart+6],
			EGLevel4:               raw[operatorStart+7],
			LevelScalingBreakPoint: raw[operatorStart+8],
			ScaleLeftDepth:         raw[operatorStart+9],
			ScaleRightDepth:        raw[operatorStart+10],

			ScaleLeftCurve:  raw[operatorStart+11],
			ScaleRightCurve: raw[operatorStart+12],

			RateScale: raw[operatorStart+13],

			AmplitudeModulationSensitivity: raw[operatorStart+14],
			KeyVelocitySensitivity:         raw[operatorStart+15],

			OutputLevel: raw[operatorStart+16],

			OscillatorMode:  raw[operatorStart+17],
			FrequencyCoarse: raw[operatorStart+18],

			FrequencyFine: raw[operatorStart+19],

			Detune: raw[operatorStart+20],
		}
		operatorStart += 21
	}

	return operators

}

// Repackage one of the 32 voices in order to send it individually
func (bank *Bank) RePackage() {

}

func (bank *Bank) DisplayVoices() error {

	log(fmt.Sprintf("Start: %X", bank.Start), nil)

	log(fmt.Sprintf("Manufacturer: %X", bank.Manufacturer), nil)
	log(fmt.Sprintf("Status and Channel: %X", bank.StatusAndChannel), nil)
	log(fmt.Sprintf("Format: %X", bank.Format), nil)
	log(fmt.Sprintf("Size: %d", bank.VoiceCount), nil)

	log(fmt.Sprintf("Voice Count: %d", bank.VoiceCount), nil)

	for i := 0; i < bank.VoiceCount; i++ {
		log(fmt.Sprintf("[%d] Name: %v", i+1, bank.Voice[i].Name), nil)

		// Start Operators
		for n, operator := range bank.Voice[i].Operators {
			log(fmt.Sprintf("		Operator %d", n+1), nil)
			log(fmt.Sprintf("			EGRate1: %.2d		EGLevel1: %.2d		ScaleLeftDepth:  %d", operator.EGRate1, operator.EGLevel1, operator.ScaleLeftDepth), nil)
			log(fmt.Sprintf("			EGRate2: %.2d		EGLevel2: %.2d		ScaleRightDepth: %d", operator.EGRate2, operator.EGLevel2, operator.ScaleRightDepth), nil)
			log(fmt.Sprintf("			EGRate3: %.2d		EGLevel3: %.2d		ScaleLeftCurve:  %d", operator.EGRate3, operator.EGLevel3, operator.ScaleLeftCurve), nil)
			log(fmt.Sprintf("			EGRate4: %.2d		EGLevel4: %.2d		ScaleRightCurve: %d", operator.EGRate4, operator.EGLevel4, operator.ScaleRightCurve), nil)
			log("", nil)

			log(fmt.Sprintf("			LevelScalingBreakPoint: %d			AmplitudeModulationSensitivity: %d", operator.LevelScalingBreakPoint, operator.AmplitudeModulationSensitivity), nil)
			log(fmt.Sprintf("			RateScale: %d					KeyVelocitySensitivity: %d", operator.RateScale, operator.KeyVelocitySensitivity), nil)
			log(fmt.Sprintf("			Detune: %d					OutputLevel: %d", operator.Detune, operator.OutputLevel), nil)
			log("", nil)

			log(fmt.Sprintf("			FrequencyCoarse: %d				OscillatorMode: %d", operator.FrequencyCoarse, operator.OscillatorMode), nil)
			log(fmt.Sprintf("			FrequencyFine: %d", operator.FrequencyFine), nil)
			log("", nil)
		}
		// End Operators

		log(fmt.Sprintf("		PitchEGRate1: %.2d		PitchEGLevel1: %.2d", bank.Voice[i].PitchEGRate1, bank.Voice[i].PitchEGLevel1), nil)
		log(fmt.Sprintf("		PitchEGRate2: %.2d		PitchEGLevel2: %.2d", bank.Voice[i].PitchEGRate2, bank.Voice[i].PitchEGLevel2), nil)
		log(fmt.Sprintf("		PitchEGRate3: %.2d		PitchEGLevel3: %.2d", bank.Voice[i].PitchEGRate3, bank.Voice[i].PitchEGLevel3), nil)
		log(fmt.Sprintf("		PitchEGRate4: %.2d		PitchEGLevel4: %.2d", bank.Voice[i].PitchEGRate4, bank.Voice[i].PitchEGLevel4), nil)
		log("", nil)

		log(fmt.Sprintf("		Algorithm: %.2d			Feedback: %.2d			OscKeySync: %.2d", bank.Voice[i].Algorithm, bank.Voice[i].Feedback, bank.Voice[i].OscKeySync), nil)
		log("", nil)

		log(fmt.Sprintf("		LfoSpeed: %.2d			LfoPitchModDepth: %.2d", bank.Voice[i].LfoSpeed, bank.Voice[i].LfoPitchModDepth), nil)
		log(fmt.Sprintf("		LfoDelay: %.2d			LfoAMDepth: %.2d", bank.Voice[i].LfoDelay, bank.Voice[i].LfoAMDepth), nil)
		log("", nil)

		log(fmt.Sprintf("		LfoSync: %.2d", bank.Voice[i].LfoSync), nil)
		log(fmt.Sprintf("		LfoWave: %.2d", bank.Voice[i].LfoWave), nil)
		log(fmt.Sprintf("		LfoPitchModSensitivity: %.2d", bank.Voice[i].LfoPitchModSensitivity), nil)
		log("", nil)

		log(fmt.Sprintf("		Transpose: %.2d", bank.Voice[i].Transpose), nil)
		log("\n\n", nil)

	}

	log(fmt.Sprintf("Checksum: %X", bank.Checksum), nil)
	log(fmt.Sprintf("End: %X", bank.End), nil)

	return nil

}

// Debug Function
////////////////..........
func dbg(kind string, err error) {
	if debug {
		if err == nil {
			fmt.Printf("### [DEBUG LOG - %s]\n\n", kind)
		} else {
			fmt.Printf("### [DEBUG ERROR - %s]: %s\n\n", kind, err)
		}
	}
}

func log(kind string, err error) {
	if err == nil {
		fmt.Printf("   %s\n", kind)
	} else {
		fmt.Printf("[ERROR - %s]: %s\n", kind, err)
	}
}
