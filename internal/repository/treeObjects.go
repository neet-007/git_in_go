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

func treeParseLeaf(reader *bytes.Reader) (*GitTreeLeaf, error) {
	modeBytes := make([]byte, 0, 6)
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("tree is malformed: could not read mode byte")
		}
		if b == ' ' {
			break
		}
		modeBytes = append(modeBytes, b)
	}

	lenMode := len(modeBytes)
	if lenMode != 5 && lenMode != 6 {
		return nil, fmt.Errorf("tree is malformed: len mode byte not 5 or 6 got %d\n", lenMode)
	}
	if lenMode == 5 {
		modeBytes = append(modeBytes, ' ')
	}

	pathBytes := []byte{}
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("tree is malformed: could not read path byte")
		}
		if b == '\x00' {
			break
		}
		pathBytes = append(pathBytes, b)
	}

	shaBytes := make([]byte, 20)
	n, err := reader.Read(shaBytes)
	if err != nil || n != 20 {
		return nil, fmt.Errorf("tree is malformed: incomplete SHA hash")
	}
	shaHex := hex.EncodeToString(shaBytes)

	return &GitTreeLeaf{
		Mode: modeBytes,
		Path: string(pathBytes),
		Sha:  shaHex,
	}, nil
}

func TreeParser(raw []byte) ([]*GitTreeLeaf, error) {
	reader := bytes.NewReader(raw)
	ret := []*GitTreeLeaf{}

	for reader.Len() > 0 {
		obj, err := treeParseLeaf(reader)
		if err != nil {
			return ret, err
		}
		ret = append(ret, obj)
	}

	return ret, nil
}

func TreeSerialize(tree *GitTree) ([]byte, error) {
	if tree == nil {
		return []byte{}, nil
	}

	//SORT THIS LIKE GIT
	ret := make([]byte, 0)

	for _, leaf := range tree.Items {
		ret = append(ret, leaf.Mode...)
		ret = append(ret, ' ')
		ret = append(ret, []byte(leaf.Path)...)
		ret = append(ret, '\x00')
		ret = append(ret, []byte(leaf.Sha)...)
	}

	return ret, nil
}
