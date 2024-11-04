package repository

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
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

	file, err := os.Open(indexFile)
	if err != nil {
		return &GitIndex{}, err
	}
	defer file.Close()

	idx := 0
	header := make([]byte, 12, 12)
	buff := make([]byte, 4, 4)
	twoBitBuff := make([]byte, 2, 2)
	tenBitBuff := make([]byte, 10, 10)

	n, err := file.Read(header)
	if err != nil {
		return &GitIndex{}, err
	}
	idx += n

	sig := header[:4]
	if string(sig) != "DIRC" {
		return &GitIndex{}, fmt.Errorf("exp sig to be DIRC got %s\n", string(sig))
	}

	version := binary.BigEndian.Uint32(header[4:8])
	if version != 2 {
		return &GitIndex{}, fmt.Errorf("exp version to be %d got %d\n", 2, version)
	}

	count := binary.BigEndian.Uint32(header[8:12])

	fmt.Printf("idx:%d\n", idx)
	entries := []GitIndexEntry{}
	for i := 0; i < int(count); i++ {
		n, err = file.ReadAt(buff, int64(idx))
		if err != nil {
			return &GitIndex{}, err
		}
		idx += n

		fmt.Printf("idx:%d\n", idx)
		cTimeS := binary.BigEndian.Uint32(buff)

		n, err = file.ReadAt(buff, int64(idx))
		if err != nil {
			return &GitIndex{}, err
		}
		idx += n
		fmt.Printf("idx:%d\n", idx)

		cTimeN := binary.BigEndian.Uint32(buff)

		n, err = file.ReadAt(buff, int64(idx))
		if err != nil {
			return &GitIndex{}, err
		}
		idx += n
		fmt.Printf("idx:%d\n", idx)

		mTimeS := binary.BigEndian.Uint32(buff)

		n, err = file.ReadAt(buff, int64(idx))
		if err != nil {
			return &GitIndex{}, err
		}
		idx += n
		fmt.Printf("idx:%d\n", idx)

		mTimeN := binary.BigEndian.Uint32(buff)

		n, err = file.ReadAt(buff, int64(idx))
		if err != nil {
			return &GitIndex{}, err
		}
		idx += n
		fmt.Printf("idx:%d\n", idx)

		dev := binary.BigEndian.Uint32(buff)

		n, err = file.ReadAt(buff, int64(idx))
		if err != nil {
			return &GitIndex{}, err
		}
		idx += n
		fmt.Printf("idx:%d\n", idx)

		ino := binary.BigEndian.Uint32(buff)

		n, err = file.ReadAt(twoBitBuff, int64(idx))
		if err != nil {
			return &GitIndex{}, err
		}
		idx += n
		fmt.Printf("idx:%d\n", idx)

		if binary.BigEndian.Uint16(twoBitBuff) != 0 {
			return &GitIndex{}, fmt.Errorf("exp unused to be %d got %d for offest %d\n", 0, binary.BigEndian.Uint16(twoBitBuff), n)
		}

		n, err = file.ReadAt(twoBitBuff, int64(idx))
		if err != nil {
			return &GitIndex{}, err
		}
		idx += n
		fmt.Printf("idx:%d\n", idx)

		mode := binary.BigEndian.Uint16(twoBitBuff)
		modeType := mode >> 12
		if modeType != 0b1000 && modeType != 0b1010 && modeType != 0b1110 {
			return &GitIndex{}, fmt.Errorf("exp mode type to be %d or %d or %d  got %d for offest %d\n", 0b1000, 0b1010, 0b1110, modeType, n)
		}
		modePerrms := mode & 0b0000000111111111

		n, err = file.ReadAt(buff, int64(idx))
		if err != nil {
			return &GitIndex{}, err
		}
		idx += n
		fmt.Printf("idx:%d\n", idx)

		uid := binary.BigEndian.Uint32(buff)

		n, err = file.ReadAt(buff, int64(idx))
		if err != nil {
			return &GitIndex{}, err
		}
		idx += n
		fmt.Printf("idx:%d\n", idx)

		gid := binary.BigEndian.Uint32(buff)

		n, err = file.ReadAt(buff, int64(idx))
		if err != nil {
			return &GitIndex{}, err
		}
		idx += n
		fmt.Printf("idx:%d\n", idx)

		fsize := binary.BigEndian.Uint32(buff)

		n, err = file.ReadAt(tenBitBuff, int64(idx))
		if err != nil {
			return &GitIndex{}, err
		}
		idx += n
		fmt.Printf("idx:%d\n", idx)

		sha := fmt.Sprintf("%040x", new(big.Int).SetBytes(tenBitBuff))

		n, err = file.ReadAt(twoBitBuff, int64(idx))
		if err != nil {
			return &GitIndex{}, err
		}
		idx += n
		fmt.Printf("idx:%d\n", idx)

		flags := binary.BigEndian.Uint16(twoBitBuff)

		flagAssumedValid := (flags & 0b1000000000000000) != 0
		flagExnteded := (flags & 0b0100000000000000) != 0

		if flagExnteded {
			return &GitIndex{}, fmt.Errorf("exp flag extended  to be false got true\n")
		}

		flagStage := flags & 0b0011000000000000

		nameLength := flags & 0b0000111111111111
		nameBuff := make([]byte, nameLength, nameLength)
		fmt.Println("bef name")
		if nameLength < 0xFFF {
			n, err = file.ReadAt(nameBuff, int64(idx))
			if err != nil {
				return &GitIndex{}, err
			}
			idx += n
			fmt.Printf("idx:%d\n", idx)

		} else {
			buffer := make([]byte, nameLength, nameLength)

			for {
				n, err := file.Read(buffer)
				if err != nil {
					if err.Error() == "EOF" {
						break
					}
					return &GitIndex{}, err
				}
				idx += n
				fmt.Printf("idx:%d\n", idx)

				nullIdx := bytes.IndexByte(buffer[:n], 0x00)

				if nullIdx != -1 {
					nameBuff = append(nameBuff, buffer[:nullIdx]...)
					idx += nullIdx + 1
					break
				} else {
					nameBuff = append(nameBuff, buffer[:n]...)
					idx += n
				}
			}
		}

		name := string(nameBuff)

		fmt.Printf("idx:%d\n", idx)
		idx = int(8 * math.Ceil(float64(idx)/8))

		entries = append(entries, GitIndexEntry{
			ModeType: modeType,
			Sha:      sha,
			Name:     name,
			CTime: fTime{
				Seconds:     cTimeS,
				Nanoseconds: cTimeN,
			},
			MTime: fTime{
				Seconds:     mTimeS,
				Nanoseconds: mTimeN,
			},
			UId:              uid,
			GId:              gid,
			Dev:              dev,
			Ino:              ino,
			ModePerms:        modePerrms,
			FSize:            fsize,
			FlagAssumedValid: flagAssumedValid,
			FlagStage:        flagStage,
		})
	}
	return &GitIndex{Entries: entries, Version: version}, nil
}
