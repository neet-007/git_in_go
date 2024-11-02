package repository

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

type GitTreeLeaf struct {
	Mode []byte
	Path string
	Sha  string
}

func treeParseLeaf(read int, reader *bytes.Reader) (int, *GitTreeLeaf, error) {
	if reader == nil {
		return 0, &GitTreeLeaf{}, fmt.Errorf("reader is nil")
	}

	spc, err := reader.ReadByte()
	if err != nil {
		return 0, &GitTreeLeaf{}, fmt.Errorf("could not read the first byte")
	}

	modeBytes := []byte{}
	for spc != ' ' {
		modeBytes = append(modeBytes, spc)
		spc, err = reader.ReadByte()
		if err != nil {
			return 0, &GitTreeLeaf{}, fmt.Errorf("tree is malformed: could not find space")
		}
	}
	read += len(modeBytes) + 1

	if len(modeBytes) != 5 && len(modeBytes) != 6 {
		return 0, &GitTreeLeaf{}, fmt.Errorf("tree is malformed: mode length must be 5 or 6, got %d", len(modeBytes))
	}
	if len(modeBytes) == 5 {
		modeBytes = append(modeBytes, ' ')
	}

	pathBytes := []byte{}
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return 0, &GitTreeLeaf{}, fmt.Errorf("tree is malformed: could not find null terminator")
		}
		if b == '\x00' {
			break
		}
		pathBytes = append(pathBytes, b)
	}
	read += len(pathBytes) + 1

	shaBytes := make([]byte, 20)
	n, err := reader.Read(shaBytes)
	if err != nil || n != 20 {
		return 0, &GitTreeLeaf{}, fmt.Errorf("tree is malformed: SHA hash is incomplete")
	}
	read += n

	shaHex := hex.EncodeToString(shaBytes)
	if len(shaHex) < 40 {
		shaHex = fmt.Sprintf("%040s", shaHex)
	}

	return read, &GitTreeLeaf{
		Mode: modeBytes,
		Path: string(pathBytes),
		Sha:  shaHex,
	}, nil
}

func TreeParser(raw []byte) ([]*GitTreeLeaf, error) {
	reader := bytes.NewReader(raw)

	lenRaw := len(raw)
	read := 0
	ret := make([]*GitTreeLeaf, 0, 0)

	for read < lenRaw {
		n, obj, err := treeParseLeaf(read, reader)
		if err != nil {
			if len(ret) == 0 {
				return []*GitTreeLeaf{}, err
			}
			return ret, err
		}

		read += n
		ret = append(ret, obj)
	}

	return ret, nil
}

// could be better impleantaion
func treeParseLeaf_(start int, raw *[]byte) (int, *GitTreeLeaf, error) {
	if raw == nil {
		return 0, &GitTreeLeaf{}, fmt.Errorf("tree is nil")
	}

	spc := bytes.IndexByte((*raw)[start:], ' ')
	if spc == -1 {
		return 0, &GitTreeLeaf{}, fmt.Errorf("tree is malformed: could not find space")
	}

	mode := make([]byte, spc)
	mode = (*raw)[start:spc]

	start += spc + 1

	lenMode := len(mode)
	if lenMode != 5 && lenMode != 6 {
		return 0, &GitTreeLeaf{}, fmt.Errorf("tree is malformed: len mode must be 5 or 6, got %d", lenMode)
	}

	if lenMode == 5 {
		mode = append(mode, ' ')
	}

	nullT := bytes.IndexByte((*raw)[start:], '\x00')
	if nullT == -1 {
		return 0, &GitTreeLeaf{}, fmt.Errorf("tree is malformed: could not find null terminator")
	}

	path := make([]byte, nullT)
	path = (*raw)[start:nullT]

	start += nullT + 1

	sha := make([]byte, 20)
	sha = (*raw)[start:]

	shaHex := hex.EncodeToString(sha)
	if len(shaHex) < 40 {
		shaHex = fmt.Sprintf("%040s", shaHex)
	}
	start += 21

	return start, &GitTreeLeaf{
		Mode: mode,
		Path: string(path),
		Sha:  shaHex,
	}, nil
}

func TreeSerialize(tree *GitTree) ([]byte, error) {
	if tree == nil {
		return []byte{}, nil
	}

	//SORT THIS LIKE GIT
	ret := make([]byte, 0)

	for _, leaf := range tree.items {
		ret = append(ret, leaf.Mode...)
		ret = append(ret, ' ')
		ret = append(ret, []byte(leaf.Path)...)
		ret = append(ret, '\x00')
		ret = append(ret, []byte(leaf.Sha)...)
	}

	return ret, nil
}
