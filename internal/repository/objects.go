package repository

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/neet-007/git_in_go/internal/utils"
)

func (repo *Repository) ObjectRead(sha string) (GitObject, error) {
	path, err := repo.RepoFile(false, "objects", sha[:2], sha[2:])

	if err != nil {
		return nil, err
	}

	res, err := utils.IsFile(path)

	if err != nil {
		return nil, err
	}

	if !res {
		return nil, fmt.Errorf("not a file")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	zlibReader, err := zlib.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("error creating zlib reader: %w", err)
	}
	defer zlibReader.Close()

	var buffer bytes.Buffer
	if _, err = io.Copy(&buffer, zlibReader); err != nil {
		return nil, fmt.Errorf("error decompressing data: %w", err)
	}

	raw := buffer.Bytes()

	x := bytes.IndexByte(raw, ' ')
	if x == -1 {
		return nil, fmt.Errorf("malformed object %s: no space found", sha)
	}
	fmtType := raw[:x]

	y := bytes.IndexByte(raw, 0)
	if y == -1 {
		return nil, fmt.Errorf("malformed object %s: no null byte found", sha)
	}

	size, err := strconv.Atoi(string(raw[x+1 : y]))
	if err != nil {
		return nil, fmt.Errorf("malformed object %s: invalid size", sha)
	}
	if size != len(raw)-y-1 {
		return nil, fmt.Errorf("malformed object %s: bad length", sha)
	}

	fmt.Printf("fmt tpye %s\n", fmtType)
	var obj GitObject
	switch string(fmtType) {
	case "commit":
		obj = &GitCommit{}
	case "tree":
		obj = &GitTree{}
	case "tag":
		obj = &GitTag{}
	case "blob":
		obj = &GitBlob{}
	default:
		fmt.Printf("heeeerer \n")
		return nil, fmt.Errorf("unknown type %s for object %s", fmtType, sha)
	}

	obj.Init(raw[y+1:])

	return obj, nil
}

func ObjectWrite(obj GitObject, repo *Repository) (string, error) {
	data, err := obj.Serialize()
	if err != nil {
		return "", err
	}

	fmtType, err := obj.GetFmt()
	if err != nil {
		return "", err
	}

	length := []byte(strconv.Itoa(len(data)))

	result := append(fmtType, ' ')
	result = append(result, length...)
	result = append(result, '\x00')
	result = append(result, data...)

	shaInterface := sha1.New()
	shaInterface.Write(result)
	sha := shaInterface.Sum(nil)

	if repo != nil {
		path, err := repo.RepoFile(true, "objects", string(sha[0:2]), string(sha[2:]))

		if err != nil {
			return "", err
		}

		if _, err := os.Stat(path); os.IsNotExist(err) {
			file, err := os.Open(path)
			if err != nil {
				return "", err
			}

			defer file.Close()

			var buf bytes.Buffer

			zlibWriter := zlib.NewWriter(&buf)

			if _, err := zlibWriter.Write(result); err != nil {
				return "", err
			}

			if err := zlibWriter.Close(); err != nil {
				return "", err
			}

			compressedData := buf.Bytes()
			file.Write(compressedData)
		}
	}

	return hex.EncodeToString(sha), nil
}

func (repo *Repository) ObjectFind(name string, fmtType string, follow bool) string {
	return name
}

func ObjectHash(file *os.File, fmtType string, repo *Repository) (string, error) {

	var data []byte
	var obj GitObject

	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	switch fmtType {
	case "\bcommit":
		{
			obj = &GitCommit{}
		}
	case "\btree":
		{
			obj = &GitTree{}

		}
	case "\btag":
		{
			obj = &GitTag{}

		}
	case "blob":
		{
			obj = &GitBlob{}

		}
	default:
		{
			return "", fmt.Errorf("Unkwon type %s", fmtType)
		}
	}

	obj.Init(data)

	sha, err := ObjectWrite(obj, repo)
	if err != nil {
		return "", fmt.Errorf("Error while writing object: %w\n", err)
	}

	return sha, nil
}
