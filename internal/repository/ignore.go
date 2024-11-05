package repository

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/neet-007/git_in_go/internal/utils"
)

type GitIgnoreLine struct {
	Raw string
	Ok  bool
}

type GitIgnore struct {
	Absolue [][]GitIgnoreLine
	Scoped  *map[string][]GitIgnoreLine
}

func (ignore *GitIgnore) Init(absolue [][]GitIgnoreLine, scoped *map[string][]GitIgnoreLine) {
	ignore.Absolue = absolue
	ignore.Scoped = scoped
}

func GitIgnoreParseLine(raw string) GitIgnoreLine {
	raw = strings.TrimSpace(raw)

	if raw == "" || strings.HasPrefix(raw, "#") {
		return GitIgnoreLine{Raw: "", Ok: false}
	}
	if strings.HasPrefix(raw, "!") {
		return GitIgnoreLine{Raw: raw[1:], Ok: false}
	}
	if strings.HasPrefix(raw, "\\") {
		return GitIgnoreLine{Raw: raw[1:], Ok: true}
	}

	return GitIgnoreLine{Raw: raw, Ok: true}
}

func GitIgnoreParseLines(lines []string) []GitIgnoreLine {
	ret := make([]GitIgnoreLine, 0, len(lines))

	for _, line := range lines {
		ignoredLine := GitIgnoreParseLine(line)
		if ignoredLine.Ok {
			ret = append(ret, ignoredLine)
		}
	}

	return ret
}

func (repo *Repository) GitIgnoreRead() (*GitIgnore, error) {
	ret := GitIgnore{}

	repoFile := filepath.Join(repo.Gitdir, "info/exclude")
	isFile, err := utils.IsFile(repoFile)
	if err == nil && isFile {
		file, err := os.Open(repoFile)
		if err == nil {
			defer file.Close()
			lines := []string{}
			buffer := make([]byte, 0)
			read := make([]byte, 1)

			for {
				_, err := file.Read(read)
				if err == io.EOF {
					if len(buffer) > 0 {
						lines = append(lines, string(buffer))
					}
					break
				}
				if err != nil {
					break
				}

				if read[0] == '\n' {
					lines = append(lines, string(buffer))
					buffer = buffer[:0]
				} else {
					buffer = append(buffer, read[0])
				}
			}

			ret.Absolue = append(ret.Absolue, GitIgnoreParseLines(lines))
		}
	}

	var configHome string
	if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
		configHome = xdgConfigHome
	} else {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return &GitIgnore{}, err
		}
		configHome = filepath.Join(userHome, ".config")
	}

	globalFile := filepath.Join(configHome, "git", "ignore")
	isFile, err = utils.IsFile(globalFile)
	if err == nil && isFile {
		file, err := os.Open(globalFile)
		if err == nil {
			defer file.Close()
			lines := []string{}
			buffer := make([]byte, 0)
			read := make([]byte, 1)

			for {
				_, err := file.Read(read)
				if err == io.EOF {
					if len(buffer) > 0 {
						lines = append(lines, string(buffer))
					}
					break
				}
				if err != nil {
					break
				}

				if read[0] == '\n' {
					lines = append(lines, string(buffer))
					buffer = buffer[:0]
				} else {
					buffer = append(buffer, read[0])
				}
			}

			ret.Absolue = append(ret.Absolue, GitIgnoreParseLines(lines))
		}
	}

	index, err := repo.IndexRead()
	if err != nil {
		return &GitIgnore{}, err
	}

	for _, e := range index.Entries {
		if e.Name != ".gitignore" && strings.HasSuffix(e.Name, "/.gitignore") {
			continue
		}
		dirName := filepath.Dir(e.Name)
		contents, err := repo.ObjectRead(e.Sha)
		if err != nil {

		}
		contentsBlob, ok := contents.(*GitBlob)
		if !ok {
			fmtType, err := contents.GetFmt()
			if err != nil {
				return &GitIgnore{}, err
			}

			return &GitIgnore{}, fmt.Errorf("exp git object to be blob got %s\n", string(fmtType))
		}

		lines := strings.Split(string(contentsBlob.BlobData), "\n")
		(*ret.Scoped)[dirName] = GitIgnoreParseLines(lines)
	}

	return &ret, nil

}

var GitIgnoreDefaultCheck = errors.New("GitIgnoreDefaultCheck")

func CheckIgnorePath(rules []GitIgnoreLine, path string) (bool, error) {
	ret := false
	err := fmt.Errorf("%w\n", GitIgnoreDefaultCheck)

	for _, rule := range rules {
		matched, err := filepath.Match(rule.Raw, path)
		if err != nil {
			return false, err
		}
		if matched {
			ret = rule.Ok
			err = nil
		}
	}

	return ret, err
}

func CheckIgnoreScoped(rules *map[string][]GitIgnoreLine, path string) (bool, error) {
	if rules == nil {
		return false, fmt.Errorf("map in nil for path:%s\n", path)
	}

	parent := filepath.Dir(path)

	for {
		if _, ok := (*rules)[parent]; ok {
			res, err := CheckIgnorePath((*rules)[parent], path)
			if err != nil && !errors.Is(err, GitIgnoreDefaultCheck) {
				return false, err
			}
			if err == nil {
				return res, err
			}
		}
		if parent == "" {
			break
		}
		parent = filepath.Dir(parent)
	}

	return false, fmt.Errorf("%w\n", GitIgnoreDefaultCheck)
}

func CheckIgnoreAbsolute(rules [][]GitIgnoreLine, path string) (bool, error) {
	//parent := filepath.Dir(path)

	for _, ruleset := range rules {
		res, err := CheckIgnorePath(ruleset, path)
		if err != nil && !errors.Is(err, GitIgnoreDefaultCheck) {
			return false, err
		}
		if err == nil {
			return res, err
		}
	}

	return false, nil
}

func CheckIgnore(rules *GitIgnore, path string) (bool, error) {
	if filepath.IsAbs(path) {
		return false, fmt.Errorf("this path is absolute path must be relative to .git dir path:%s\n", path)
	}

	res, err := CheckIgnoreScoped(rules.Scoped, path)
	if err != nil && !errors.Is(err, GitIgnoreDefaultCheck) {
		return false, err
	}
	if err == nil {
		return res, err
	}

	return CheckIgnoreAbsolute(rules.Absolue, path)
}
