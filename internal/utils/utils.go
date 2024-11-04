package utils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/neet-007/git_in_go/internal/sharedtypes"
)

func IsFile(path string) (bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, fmt.Errorf("path does not exist: %w", err)
		}
		return false, err
	}

	// Check if the path is within .git/refs (valid Git reference)
	if strings.Contains(path, filepath.Join(".git", "refs")) {
		return true, nil
	}

	// Check if it's a regular file
	return info.Mode().IsRegular(), nil
}

func RealPath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return "", err
	}

	if info.Mode()&os.ModeSymlink != 0 {
		resolvedPath, err := os.Readlink(absPath)
		if err != nil {
			return "", err
		}
		return resolvedPath, nil
	}

	return absPath, nil
}

// -IMPORTANT  MAKE THIS ORDER DICT IMPORTANT
func KvlmParser(raw *[]byte, start int, parsed *sharedtypes.Kvlm) (*sharedtypes.Kvlm, error) {
	if raw == nil || parsed == nil {
		parsed = &sharedtypes.Kvlm{}
	}

	spc := bytes.IndexByte((*raw)[start:], ' ') + start
	nl := bytes.IndexByte((*raw)[start:], '\n') + start

	if spc < 0 || nl < spc {
		if start != nl {
			return nil, fmt.Errorf("final message reached but nl != start nl:%d start:%d\n", nl, start)
		}

		(*parsed)[""] = append((*parsed)[""], (*raw)[start+1:])

		return parsed, nil
	}

	key := (*raw)[start:spc]

	end := start
	for {
		end = bytes.IndexByte((*raw)[start:], '\n') + start
		if end == -1 || end+1 >= len(*raw) || (*raw)[end+1] != ' ' {
			break
		}
	}

	val, ok := (*parsed)[string(key)]
	if !ok {
		(*parsed)[string(key)] = [][]byte{(*raw)[spc+1 : end]}
	} else {
		(*parsed)[string(key)] = append(val, (*raw)[spc+1:end])
	}

	return KvlmParser(raw, end+1, parsed)
}

func KvlmSerialize(kvlm *sharedtypes.Kvlm) ([]byte, error) {
	if kvlm == nil {
		return []byte{}, fmt.Errorf("nil kvlm passed\n")
	}
	ret := []byte{}

	for key := range *kvlm {
		if key == "" {
			continue
		}

		val, ok := (*kvlm)[key]
		if !ok {
			return []byte{}, fmt.Errorf("key does not have value in kvml key:%s\n", key)
		}
		ret = append(ret, []byte(key)...)
		ret = append(ret, ' ')
		for _, b := range val {
			ret = append(ret, b...)
		}
		ret = append(ret, '\n')
	}

	val, ok := (*kvlm)[""]
	if !ok {
		return []byte{}, fmt.Errorf("key does not have value in kvml key: FINAL MESSAGE KEY\n")
	}
	ret = append(ret, '\n')
	for _, b := range val {
		ret = append(ret, b...)
	}
	ret = append(ret, '\n')

	return ret, nil
}
