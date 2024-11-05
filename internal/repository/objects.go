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
	fmt.Printf("sha:%s\n", sha)
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
	sha := hex.EncodeToString(shaInterface.Sum(nil))

	if repo != nil {
		path, err := repo.RepoFile(true, "objects", sha[0:2], sha[2:])

		if err != nil {
			return "", err
		}

		println("i me heeeeerer")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			file, err := os.Create(path)
			fmt.Printf("file path:%s\n", path)
			if err != nil {
				println("this errrro")
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

	return sha, nil
}

func (repo *Repository) ObjectFind(name string, fmtType string, follow bool) (string, error) {
	/*
		fmtType: defailt val is ""
		follow: default val is true
	*/
	sha, err := repo.ObjectResolve(name)
	if err != nil {
		return "", fmt.Errorf("for name:%s, fmtType:%s, follow:%v error:%w\n", name, fmtType, follow, err)
	}

	if len(sha) != 1 {
		return "", fmt.Errorf("sha candidates size is not 1 got %d found for name:%s, fmtType:%s, follow:%v\n", len(sha), name, fmtType, follow)
	}

	shaStr := sha[0]

	if fmtType == "" {
		return shaStr, nil
	}

	for {
		obj, err := repo.ObjectRead(shaStr)
		if err != nil {
			return "", fmt.Errorf("for name:%s, fmtType:%s, follow:%v error:%w\n", name, fmtType, follow, err)
		}

		objFmt, err := obj.GetFmt()
		if err != nil {
			return "", fmt.Errorf("for name:%s, fmtType:%s, follow:%v error:%w\n", name, fmtType, follow, err)
		}

		if string(objFmt) == fmtType {
			break
		}

		if !follow {
			return "", fmt.Errorf("not follow name:%s, fmtType:%s, follow:%v error:%w\n", name, fmtType, follow)
		}

		if fmtType == "tag" {
			objTag, ok := obj.(*GitTag)
			if !ok {
				return "", fmt.Errorf("could not convert obj to GitTag for name:%s, fmtType:%s, follow:%v\n", name, fmtType, follow)
			}

			var combined []byte
			for _, slice := range (*(*objTag.Kvlm).Map)["object"] {
				combined = append(combined, slice...)
			}

			shaStr = string(combined)
		} else if string(objFmt) == "commit" && fmtType == "tree" {
			objCommit, ok := obj.(*GitCommit)
			if !ok {
				return "", fmt.Errorf("could not convert obj to GitCommit for name:%s, fmtType:%s, follow:%v\n", name, fmtType, follow)
			}

			var combined []byte
			for _, slice := range (*(*&objCommit.Kvlm).Map)["tree"] {
				combined = append(combined, slice...)
			}

			fmt.Printf("combinde:%s\n", string(combined))
			shaStr = string(combined)
		} else {
			return "", fmt.Errorf("last case name:%s, fmtType:%s, follow:%v error:%w\n", name, fmtType, follow)
		}
	}

	return shaStr, nil
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
