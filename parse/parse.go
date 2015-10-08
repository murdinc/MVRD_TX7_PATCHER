package parse

import (
	"fmt"
	"os"
)

type Bank struct {
	FileName         string
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

func New(fileName string) (Bank, error) {

	f, err := os.Open(fileName)
	fi, err := f.Stat()
	if err != nil {
		log("DisplayPatches - Error opening bank file", err)
		return Bank{}, err
	}

	fileSize := fi.Size()
	sysexFile := make([]byte, fileSize)

	// Read in all the bytes
	n, err := f.Read(sysexFile)
	if err != nil {
		log("DisplayPatches - Error reading bank file", err)
		return Bank{}, err
	}
	log(fmt.Sprintf("DisplayPatches - reading %d bytes from bank file.", n), nil)
	log(fmt.Sprintf("DisplayPatches - [%s] is %d bytes long", fileName, fileSize), nil)

	bank := Bank{Raw: sysexFile}

	err = bank.Parse()
	if err != nil {
		log("DisplayPatches - Error parsing bank file", err)
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

	// Only if this is 32 Voices
	bank.doBulkVoices()

	bank.Checksum = bank.Raw[bank.Size+6]
	bank.End = bank.Raw[bank.Size+7] //F7

	return nil
}

func (bank *Bank) doBulkVoices() {
	bank.Voice = make([]Voice, 32)

	voiceStart := 6

	for i := 0; i < 32; i++ {

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

			Algorithm: bank.Raw[voiceStart+110], // 5
			//unsigned : : 3,

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

func (bank *Bank) DisplayPatches() error {

	log(fmt.Sprintf("Start: %X", bank.Start), nil)

	log(fmt.Sprintf("Manufacturer: %X", bank.Manufacturer), nil)
	log(fmt.Sprintf("Status and Channel: %X", bank.StatusAndChannel), nil)
	log(fmt.Sprintf("Format: %X", bank.Format), nil)
	log(fmt.Sprintf("Size: %d", bank.Size), nil)

	for i := 0; i < 32; i++ {
		log(fmt.Sprintf("[%d] Name: %v", i+1, bank.Voice[i].Name), nil)

		// Start Operators
		for n, operator := range bank.Voice[i].Operators {
			log(fmt.Sprintf("       Operator %d", n+1), nil)
			log(fmt.Sprintf("         EGRate1: %d", operator.EGRate1), nil)
			log(fmt.Sprintf("         EGRate2: %d", operator.EGRate2), nil)
			log(fmt.Sprintf("         EGRate3: %d", operator.EGRate3), nil)
			log(fmt.Sprintf("         EGRate4: %d", operator.EGRate4), nil)
			log("", nil)

			log(fmt.Sprintf("         EGLevel1: %d", operator.EGLevel1), nil)
			log(fmt.Sprintf("         EGLevel2: %d", operator.EGLevel2), nil)
			log(fmt.Sprintf("         EGLevel3: %d", operator.EGLevel3), nil)
			log(fmt.Sprintf("         EGLevel4: %d", operator.EGLevel4), nil)
			log("", nil)

			log(fmt.Sprintf("         LevelScalingBreakPoint: %d ", operator.LevelScalingBreakPoint), nil)
			log(fmt.Sprintf("         ScaleLeftDepth: %d ", operator.ScaleLeftDepth), nil)
			log(fmt.Sprintf("         ScaleRightDepth: %d", operator.ScaleRightDepth), nil)
			log(fmt.Sprintf("         ScaleLeftCurve: %d ", operator.ScaleLeftCurve), nil)
			log(fmt.Sprintf("         ScaleRightCurve: %d", operator.ScaleRightCurve), nil)
			log("", nil)

			log(fmt.Sprintf("         RateScale: %d", operator.RateScale), nil)
			log(fmt.Sprintf("         Detune: %d ", operator.Detune), nil)
			log("", nil)

			log(fmt.Sprintf("         AmplitudeModulationSensitivity: %d ", operator.AmplitudeModulationSensitivity), nil)
			log("", nil)

			log(fmt.Sprintf("         KeyVelocitySensitivity: %d ", operator.KeyVelocitySensitivity), nil)
			log("", nil)

			log(fmt.Sprintf("         OutputLevel: %d", operator.OutputLevel), nil)
			log("", nil)

			log(fmt.Sprintf("         OscillatorMode: %d ", operator.OscillatorMode), nil)
			log("", nil)

			log(fmt.Sprintf("         FrequencyCoarse: %d", operator.FrequencyCoarse), nil)
			log(fmt.Sprintf("         FrequencyFine: %d", operator.FrequencyFine), nil)
			log("", nil)
		}
		// End Operators

		log(fmt.Sprintf("     PitchEGRate1: %d", bank.Voice[i].PitchEGRate1), nil)
		log(fmt.Sprintf("     PitchEGRate2: %d", bank.Voice[i].PitchEGRate2), nil)
		log(fmt.Sprintf("     PitchEGRate3: %d", bank.Voice[i].PitchEGRate3), nil)
		log(fmt.Sprintf("     PitchEGRate4: %d", bank.Voice[i].PitchEGRate4), nil)
		log("", nil)

		log(fmt.Sprintf("     PitchEGLevel1: %d", bank.Voice[i].PitchEGLevel1), nil)
		log(fmt.Sprintf("     PitchEGLevel2: %d", bank.Voice[i].PitchEGLevel2), nil)
		log(fmt.Sprintf("     PitchEGLevel3: %d", bank.Voice[i].PitchEGLevel3), nil)
		log(fmt.Sprintf("     PitchEGLevel4: %d", bank.Voice[i].PitchEGLevel4), nil)
		log("", nil)

		log(fmt.Sprintf("     Algorithm: %d", bank.Voice[i].Algorithm), nil)
		log("", nil)

		log(fmt.Sprintf("     Feedback: %d", bank.Voice[i].Feedback), nil)
		log(fmt.Sprintf("     OscKeySync: %d", bank.Voice[i].OscKeySync), nil)
		log("", nil)

		log(fmt.Sprintf("     LfoSpeed: %d", bank.Voice[i].LfoSpeed), nil)
		log(fmt.Sprintf("     LfoDelay: %d", bank.Voice[i].LfoDelay), nil)
		log("", nil)

		log(fmt.Sprintf("     LfoPitchModDepth: %d", bank.Voice[i].LfoPitchModDepth), nil)
		log(fmt.Sprintf("     LfoAMDepth: %d", bank.Voice[i].LfoAMDepth), nil)
		log("", nil)

		log(fmt.Sprintf("     LfoSync: %d", bank.Voice[i].LfoSync), nil)
		log(fmt.Sprintf("     LfoWave: %d", bank.Voice[i].LfoWave), nil)
		log(fmt.Sprintf("     LfoPitchModSensitivity: %d", bank.Voice[i].LfoPitchModSensitivity), nil)
		log("", nil)

		log(fmt.Sprintf("     Transpose: %d", bank.Voice[i].Transpose), nil)
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
