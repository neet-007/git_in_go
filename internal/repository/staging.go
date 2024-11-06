package repository

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
)

type fTime struct {
	Seconds     uint32
	Nanoseconds uint32
}

/*
	type GitIndexEntry struct {
		ModeType         int
		Sha              string
		Name             string
		CTime            fTime
		MTime            fTime
		UId              int
		GId              int
		Dev              int
		Ino              int
		ModePerms        int
		FSize            int
		FlagStage        int
		FlagAssumedValid bool
	}
*/
type GitIndexEntry struct {
	Sha              string
	Name             string
	CTime            fTime
	MTime            fTime
	ModeType         uint16
	UId              uint32
	GId              uint32
	Dev              uint32
	Ino              uint32
	ModePerms        uint16
	FSize            uint32
	FlagStage        uint16
	FlagAssumedValid bool
}

type GitIndex struct {
	Version uint32
	Entries []GitIndexEntry
}

func (entry *GitIndexEntry) Init(ModeType uint16, CTime fTime, MTime fTime, Sha string,
	Name string, UId uint32, GId uint32, Dev uint32, Ino uint32, ModePerms uint16, FSize uint32,
	FlagAssumedValid bool, FlagStage uint16) {

	entry.ModeType = ModeType
	entry.CTime = CTime
	entry.MTime = MTime
	entry.Dev = Dev
	entry.UId = UId
	entry.GId = GId
	entry.Sha = Sha
	entry.Name = Name
	entry.Ino = Ino
	entry.ModePerms = ModePerms
	entry.FSize = FSize
	entry.FlagAssumedValid = FlagAssumedValid
	entry.FlagStage = FlagStage
}

func (index *GitIndex) Init(version uint32, entries []GitIndexEntry) {
	/*
		version: default val is 2
	*/
	if entries == nil {
		entries = []GitIndexEntry{}
	}

	index.Version = version
	index.Entries = entries
}

func (repo *Repository) IndexRead() (*GitIndex, error) {
	indexFile, err := repo.RepoFile(false, "index")
	if err != nil {
		return &GitIndex{}, err
	}
	raw, err := os.ReadFile(indexFile)
	if err != nil {
		return nil, err
	}

	if len(raw) < 12 {
		return nil, errors.New("index file too small")
	}

	header := raw[:12]
	if !bytes.Equal(header[:4], []byte("DIRC")) {
		return nil, errors.New("unexpected signature")
	}

	version := binary.BigEndian.Uint32(header[4:8])
	if version != 2 {
		return nil, fmt.Errorf("unsupported version %d", version)
	}

	count := binary.BigEndian.Uint32(header[8:12])
	entries := []GitIndexEntry{}
	content := raw[12:]

	idx := 0
	for i := 0; i < int(count); i++ {
		if idx+62 > len(content) {
			return nil, errors.New("unexpected end of data")
		}

		cTimeSec := binary.BigEndian.Uint32(content[idx : idx+4])
		cTimeNSec := binary.BigEndian.Uint32(content[idx+4 : idx+8])
		mTimeSec := binary.BigEndian.Uint32(content[idx+8 : idx+12])
		mTimeNSec := binary.BigEndian.Uint32(content[idx+12 : idx+16])
		dev := binary.BigEndian.Uint32(content[idx+16 : idx+20])
		ino := binary.BigEndian.Uint32(content[idx+20 : idx+24])
		unused := binary.BigEndian.Uint16(content[idx+24 : idx+26])

		if unused != 0 {
			return nil, fmt.Errorf("unexpected value in unused field: %d", unused)
		}

		mode := binary.BigEndian.Uint16(content[idx+26 : idx+28])
		modeType := mode >> 12
		if modeType != 0b1000 && modeType != 0b1010 && modeType != 0b1110 {
			return nil, fmt.Errorf("unexpected mode type: %d", modeType)
		}
		modePerms := mode & 0b0000000111111111
		uid := binary.BigEndian.Uint32(content[idx+28 : idx+32])
		gid := binary.BigEndian.Uint32(content[idx+32 : idx+36])
		fSize := binary.BigEndian.Uint32(content[idx+36 : idx+40])
		sha := fmt.Sprintf("%040x", content[idx+40:idx+60])
		flags := binary.BigEndian.Uint16(content[idx+60 : idx+62])

		flagAssumeValid := (flags & 0b1000000000000000) != 0
		flagStage := flags & 0b0011000000000000
		nameLength := flags & 0b0000111111111111

		idx += 62

		var name []byte
		if nameLength < 0xFFF {
			name = content[idx : idx+int(nameLength)]
			idx += int(nameLength) + 1 // skip null terminator
		} else {
			nullIdx := bytes.IndexByte(content[idx:], 0x00)
			if nullIdx == -1 {
				return nil, errors.New("name terminator not found")
			}
			name = content[idx : idx+nullIdx]
			idx += nullIdx + 1
		}

		nameStr := string(name)
		idx = int(8 * math.Ceil(float64(idx)/8))

		entry := GitIndexEntry{

			ModeType: modeType,
			Sha:      sha,
			Name:     nameStr,
			CTime: fTime{
				Seconds:     cTimeSec,
				Nanoseconds: cTimeNSec,
			},
			MTime: fTime{
				Seconds:     mTimeSec,
				Nanoseconds: mTimeNSec,
			},
			UId:              uid,
			GId:              gid,
			Dev:              dev,
			Ino:              ino,
			ModePerms:        modePerms,
			FSize:            fSize,
			FlagAssumedValid: flagAssumeValid,
			FlagStage:        flagStage,
		}

		entries = append(entries, entry)
	}

	return &GitIndex{Version: version, Entries: entries}, nil
}

func (repo *Repository) IndexWrite(index *GitIndex) error {
	indexFile, err := repo.RepoFile(false, "index")
	if err != nil {
		return err
	}

	file, err := os.OpenFile(indexFile, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(0, 0)
	if err != nil {
		return err
	}

	dirc := make([]byte, 4, 4)
	for _, c := range "DIRC" {
		dirc = append(dirc, byte(c))
	}

	_, err = file.Write(dirc)
	if err != nil {
		return err
	}

	_, err = file.Write(binary.BigEndian.AppendUint32([]byte{}, index.Version))
	if err != nil {
		return err
	}

	_, err = file.Write(binary.BigEndian.AppendUint32([]byte{}, uint32(len(index.Entries))))
	if err != nil {
		return err
	}

	idx := 0
	for _, e := range index.Entries {
		_, err = file.Write(binary.BigEndian.AppendUint32([]byte{}, e.CTime.Seconds))
		if err != nil {
			return err
		}

		_, err = file.Write(binary.BigEndian.AppendUint32([]byte{}, e.CTime.Nanoseconds))
		if err != nil {
			return err
		}

		_, err = file.Write(binary.BigEndian.AppendUint32([]byte{}, e.MTime.Seconds))
		if err != nil {
			return err
		}

		_, err = file.Write(binary.BigEndian.AppendUint32([]byte{}, e.MTime.Nanoseconds))
		if err != nil {
			return err
		}

		_, err = file.Write(binary.BigEndian.AppendUint32([]byte{}, e.Dev))
		if err != nil {
			return err
		}

		_, err = file.Write(binary.BigEndian.AppendUint32([]byte{}, e.Ino))
		if err != nil {
			return err
		}

		mode := (e.ModeType << 12) | e.ModePerms
		_, err = file.Write(binary.BigEndian.AppendUint16([]byte{}, mode))
		if err != nil {
			return err
		}

		_, err = file.Write(binary.BigEndian.AppendUint32([]byte{}, e.UId))
		if err != nil {
			return err
		}

		_, err = file.Write(binary.BigEndian.AppendUint32([]byte{}, e.GId))
		if err != nil {
			return err
		}

		_, err = file.Write(binary.BigEndian.AppendUint32([]byte{}, e.FSize))
		if err != nil {
			return err
		}

		shaBytes := make([]byte, len(e.Sha))
		for _, c := range e.Sha {
			shaBytes = append(shaBytes, byte(c))
		}
		_, err = file.Write(shaBytes)
		if err != nil {
			return err
		}

		var flagAssumedValid uint16
		if e.FlagAssumedValid {
			flagAssumedValid = 0x1 << 15
		} else {
			flagAssumedValid = 0
		}

		lenNameBytes := len(e.Name)
		nameBytes := make([]byte, lenNameBytes)
		for _, c := range e.Name {
			nameBytes = append(nameBytes, byte(c))
		}

		var nameLength int = lenNameBytes
		if lenNameBytes >= 0xFFF {
			nameLength = 0xFFF
		}

		file.Write(binary.BigEndian.AppendUint16([]byte{}, (flagAssumedValid | e.FlagStage | uint16(nameLength))))

		file.Write(nameBytes)
		_, err = file.Write([]byte{0})

		idx += 62 + len(nameBytes) + 1

		if idx%8 != 0 {
			pad := 8 - (idx % 8)
			paddingBytes := make([]byte, pad)

			_, err = file.Write(paddingBytes)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
