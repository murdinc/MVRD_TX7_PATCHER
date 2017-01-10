package parse

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mitchellh/hashstructure"
	"github.com/murdinc/terminal"
)

type Library struct {
	//Banks      []Bank
	voices    []Voice
	FileCount int
	//VoiceCount int
	Duplicates    int
	FolderName    string
	HashMap       map[uint64]string
	SearchStr     string
	searchResults []Voice
}

type Bank struct {
	FileName         string
	VoiceCount       int
	Voices           []Voice
	Raw              []byte
	Start            byte
	Manufacturer     byte
	StatusAndChannel byte
	Format           byte
	Size             int16
	Checksum         byte
	End              byte
	HashMap          *map[uint64]string
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
	BankFileName           string `hash:"ignore"`
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

	terminal.Information("Reading sysex folder...")

	files := []string{}
	filepath.Walk(foldername, func(path string, f os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})

	//

	banks := make([]Bank, 0)
	voices := make([]Voice, 0)
	duplicates := 0
	hashMap := make(map[uint64]string)

	for _, file := range files {
		if strings.HasSuffix(strings.ToLower(file), ".syx") {
			//log(fmt.Sprintf("Scanning File: [%s]", file.Name()), nil)

			bank, bankDuplicates, err := Open(file, &hashMap)
			duplicates += bankDuplicates

			if err == nil {
				banks = append(banks, bank)
				voices = append(voices, bank.Voices...)
			}
		}
	}

	foldername = strings.TrimPrefix(foldername, "./")

	library := Library{voices: voices, FileCount: len(banks), FolderName: foldername, Duplicates: duplicates}

	log(fmt.Sprintf("Files:  [ %d ]", library.FileCount), nil)

	log(fmt.Sprintf("Voices: [ %d ]", library.VoiceCount()), nil)

	return library, nil

}

func (l *Library) Voices() []Voice {

	if len(l.SearchStr) > 0 {
		l.searchResults = []Voice{}

		term := regexp.MustCompile(strings.ToLower(l.SearchStr))

	Loop:
		for _, voice := range l.voices {

			if term.MatchString(strings.ToLower(voice.Name)) {
				l.searchResults = append(l.searchResults, voice)
				continue Loop
			}
		}

		//if len(l.searchResults) > 0 {
		return l.searchResults
		//}
	}

	return l.voices
}

func (l *Library) Search(str string) {
	l.SearchStr = str

	return
}

func (l *Library) VoiceCount() int {
	return len(l.Voices())
}

func (l *Library) BuildSysex(voiceIndex int) []byte {

	voice := l.Voices()[voiceIndex]

	sysex := []byte{0xF0, 0x43, 0x00, 0x00, 0x01, 0x1B} // data1 - data155 --- checksum, 0xF7

	for _, oper := range voice.Operators {
		sysex = append(sysex, []byte{
			oper.EGRate1, oper.EGRate2, oper.EGRate3, oper.EGRate4, oper.EGLevel1, oper.EGLevel2, oper.EGLevel3, oper.EGLevel4,
			oper.LevelScalingBreakPoint, oper.ScaleLeftDepth, oper.ScaleRightDepth, oper.ScaleLeftCurve, oper.ScaleRightCurve,
			oper.RateScale, oper.AmplitudeModulationSensitivity, oper.KeyVelocitySensitivity, oper.OutputLevel, oper.OscillatorMode,
			oper.FrequencyCoarse, oper.FrequencyFine, oper.Detune}...)
	}

	sysex = append(sysex, []byte{
		voice.PitchEGRate1, voice.PitchEGRate2, voice.PitchEGRate3, voice.PitchEGRate4,
		voice.PitchEGLevel1, voice.PitchEGLevel2, voice.PitchEGLevel3, voice.PitchEGLevel4,
		voice.Algorithm, voice.Feedback, voice.OscKeySync, voice.LfoSpeed, voice.LfoDelay,
		voice.LfoPitchModDepth, voice.LfoAMDepth, voice.LfoSync, voice.LfoWave, voice.LfoPitchModSensitivity,
		voice.Transpose}...)

	sysex = append(sysex, []byte(voice.Name)...)

	sysex = append(sysex, checksum(sysex[6:]), 0xF7)

	return sysex
}

func checksum(block []byte) byte {
	crc := byte(0x00)
	for i := 0; i < len(block); i++ {
		crc += (block[i] & 0x7F)
	}
	crc = (^crc) + 1
	crc &= 0x7F
	return crc

}

func (l *Library) Length() int {

	return l.FileCount

}

func (l *Library) DisplayVoiceNames() {

	i := 0

	for _, voice := range l.Voices() {
		i++
		log(fmt.Sprintf("[# %d]		Voice Name: [%s]		File Name: [%s] ", i, voice.Name, voice.BankFileName), nil)
	}

}

func Open(fileName string, hashMap *map[uint64]string) (Bank, int, error) {

	f, err := os.Open(fileName)
	fi, err := f.Stat()
	if err != nil {
		log(fmt.Sprintf("Open - Error opening bank file %s", fileName), err)
		return Bank{}, 0, err
	}
	defer f.Close()

	fileSize := fi.Size()
	sysexFile := make([]byte, fileSize)

	// Read in all the bytes
	_, err = f.Read(sysexFile)
	if err != nil {
		log("Open - Error reading bank file", err)
		return Bank{}, 0, err
	}

	bank := Bank{Raw: sysexFile, FileName: fileName, HashMap: hashMap}

	duplicates, err := bank.Parse()
	if err != nil {
		log("Open - Error parsing bank file", err)
		return Bank{}, 0, err
	}

	return bank, duplicates, nil
}

func New(raw []byte) (Bank, error) {
	bank := Bank{Raw: raw}

	_, err := bank.Parse()
	if err != nil {
		log("New - Error parsing bank file", err)
		return Bank{}, err
	}

	return bank, nil
}

func (bank *Bank) Parse() (int, error) {
	bank.Start = bank.Raw[0] //F0
	bank.Manufacturer = bank.Raw[1]
	bank.StatusAndChannel = bank.Raw[2]
	bank.Format = bank.Raw[3]
	bank.Size = int16((int16(bank.Raw[4]) << 7) | int16(bank.Raw[5]))

	duplicates := 0

	if bank.Start != 0xF0 {
		//terminal.ErrorLine(fmt.Sprintf("Unknown Start / Status Byte: 0x%.2X, skipping: %v", bank.Start, bank.FileName))
		return duplicates, nil
	}

	if (bank.Format != 0x00) && (bank.Format != 0x09) {
		//terminal.ErrorLine(fmt.Sprintf("Unknown Bank Format: 0x%.2X, skipping: %v", bank.Format, bank.FileName))
		return duplicates, nil
	}

	if len(bank.Raw) < int(bank.Size+8) {
		//terminal.ErrorLine(fmt.Sprintf("Unexpectedly truncated File Size. Bank Format: 0x%.2X, FILE: %v", bank.Format, bank.FileName))
		return duplicates, nil
	}

	switch bank.Format {
	case 0x00:
		bank.VoiceCount = 1
		duplicates = bank.doSingleVoice()
		bank.VoiceCount -= duplicates
		//bank.Size = 155

	case 0x09:
		bank.VoiceCount = 32
		duplicates = bank.doBulkVoices()
		bank.VoiceCount -= duplicates
		//bank.Size = 4096

	default:
		terminal.ErrorLine(fmt.Sprintf("Unknown Bank Format: 0x%.2X, skipping: %v", bank.Format, bank.FileName))
		return duplicates, nil
	}

	//terminal.Information(fmt.Sprintf("Bank Format: 0x%.2X, Bank Size: %d file: %v", bank.Format, bank.Size, bank.FileName))

	bank.Checksum = bank.Raw[bank.Size+6]
	bank.End = bank.Raw[bank.Size+7]

	return duplicates, nil
}

func (bank *Bank) doSingleVoice() int {

	voiceStart := 6
	duplicates := 0

	bank.Voices = make([]Voice, 1)

	voice := Voice{

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

		Name:         string(bank.Raw[voiceStart+145 : voiceStart+155]),
		BankFileName: bank.FileName,
	}

	voiceHash, _ := hashstructure.Hash(voice, nil)

	if _, ok := (*bank.HashMap)[voiceHash]; ok {
		//terminal.Notice("Duplicate found!	-	" + existingName)
		duplicates++

	} else {
		bank.Voices[0] = voice
		(*bank.HashMap)[voiceHash] = voice.Name
	}

	return duplicates
}

func (bank *Bank) doBulkVoices() int {
	bank.Voices = make([]Voice, 32)

	voiceStart := 6
	duplicates := 0

	for i := 0; i < bank.VoiceCount; i++ {

		end := voiceStart + 128

		if len(bank.Raw[voiceStart+118:]) < 10 {
			end = len(bank.Raw[voiceStart+118:]) + voiceStart + 118
		}

		voice := Voice{

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

			Name:         string(bank.Raw[voiceStart+118 : end]),
			BankFileName: bank.FileName,
		}

		voiceHash, _ := hashstructure.Hash(voice, nil)

		if _, ok := (*bank.HashMap)[voiceHash]; ok {
			duplicates++
			bank.Voices = bank.Voices[:len(bank.Voices)-1]

		} else {
			bank.Voices[i-duplicates] = voice
			(*bank.HashMap)[voiceHash] = voice.Name
		}

		voiceStart += 128
	}

	return duplicates

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
		log(fmt.Sprintf("[%d] Name: %v", i+1, bank.Voices[i].Name), nil)

		// Start Operators
		for n, operator := range bank.Voices[i].Operators {
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

		log(fmt.Sprintf("		PitchEGRate1: %.2d		PitchEGLevel1: %.2d", bank.Voices[i].PitchEGRate1, bank.Voices[i].PitchEGLevel1), nil)
		log(fmt.Sprintf("		PitchEGRate2: %.2d		PitchEGLevel2: %.2d", bank.Voices[i].PitchEGRate2, bank.Voices[i].PitchEGLevel2), nil)
		log(fmt.Sprintf("		PitchEGRate3: %.2d		PitchEGLevel3: %.2d", bank.Voices[i].PitchEGRate3, bank.Voices[i].PitchEGLevel3), nil)
		log(fmt.Sprintf("		PitchEGRate4: %.2d		PitchEGLevel4: %.2d", bank.Voices[i].PitchEGRate4, bank.Voices[i].PitchEGLevel4), nil)
		log("", nil)

		log(fmt.Sprintf("		Algorithm: %.2d			Feedback: %.2d			OscKeySync: %.2d", bank.Voices[i].Algorithm, bank.Voices[i].Feedback, bank.Voices[i].OscKeySync), nil)
		log("", nil)

		log(fmt.Sprintf("		LfoSpeed: %.2d			LfoPitchModDepth: %.2d", bank.Voices[i].LfoSpeed, bank.Voices[i].LfoPitchModDepth), nil)
		log(fmt.Sprintf("		LfoDelay: %.2d			LfoAMDepth: %.2d", bank.Voices[i].LfoDelay, bank.Voices[i].LfoAMDepth), nil)
		log("", nil)

		log(fmt.Sprintf("		LfoSync: %.2d", bank.Voices[i].LfoSync), nil)
		log(fmt.Sprintf("		LfoWave: %.2d", bank.Voices[i].LfoWave), nil)
		log(fmt.Sprintf("		LfoPitchModSensitivity: %.2d", bank.Voices[i].LfoPitchModSensitivity), nil)
		log("", nil)

		log(fmt.Sprintf("		Transpose: %.2d", bank.Voices[i].Transpose), nil)
		log("\n\n", nil)

	}

	log(fmt.Sprintf("Checksum: %X", bank.Checksum), nil)

	dataRange := bank.Raw[6:161]
	checkSum := checksum(dataRange)
	log(fmt.Sprintf("Calculated Checksum: %X", checkSum), nil)
	log(fmt.Sprintf("Calculated Checksum Length: %d", len(dataRange)), nil)

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
