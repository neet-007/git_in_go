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
	startPos, _ := reader.Seek(0, 1)
	raw := make([]byte, reader.Len())
	reader.ReadAt(raw, startPos)

	spc := bytes.IndexByte(raw, ' ')
	if spc == -1 {
		return 0, &GitTreeLeaf{}, fmt.Errorf("tree is malformed: could not find space")
	}

	mode := make([]byte, spc)
	n, err := reader.Read(mode)
	if err != nil {
		return 0, &GitTreeLeaf{}, err
	}

	read += n + 1

	lenMode := len(mode)
	if lenMode != 5 && lenMode != 6 {
		return 0, &GitTreeLeaf{}, fmt.Errorf("tree is malformed: len mode must be 5 or 6, got %d", lenMode)
	}

	if lenMode == 5 {
		mode = append(mode, ' ')
	}

	nullT := bytes.IndexByte(raw[read:], '\x00')
	if nullT == -1 {
		return 0, &GitTreeLeaf{}, fmt.Errorf("tree is malformed: could not find null terminator")
	}

	path := make([]byte, nullT)
	n, err = reader.Read(path)
	if err != nil {
		return 0, &GitTreeLeaf{}, err
	}

	read += n + 1

	sha := make([]byte, 20)
	n, err = reader.Read(sha)
	if err != nil {
		return 0, &GitTreeLeaf{}, err
	}

	shaHex := hex.EncodeToString(sha)
	if len(shaHex) < 40 {
		shaHex = fmt.Sprintf("%040s", shaHex)
	}
	read += n + 1

	return read, &GitTreeLeaf{
		Mode: mode,
		Path: string(path),
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
				fmt.Println("errrrrrrrrrr?")
				return []*GitTreeLeaf{}, err
			}
			return ret, err
		}

		read += n
		fmt.Printf("read:%d vs len:%d\n", n, lenRaw)
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
