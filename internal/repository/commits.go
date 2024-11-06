package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/neet-007/git_in_go/internal/sharedtypes"
	"github.com/neet-007/git_in_go/internal/utils"
	"gopkg.in/ini.v1"
)

func LogGraphviz(repo *Repository, sha string, seen *map[string]byte) error {
	if seen == nil {
		return fmt.Errorf("seen map is nil for commit:%s", sha)
	}

	if _, ok := (*seen)[sha]; ok {
		return nil
	}

	(*seen)[sha] = ' '
	commit, err := repo.ObjectRead(sha)
	if err != nil {
		return fmt.Errorf("could not find repo in loggin for sha:%s", sha)
	}

	commitFmt, err := commit.GetFmt()
	if err != nil {
		return fmt.Errorf("error in fmt for commit with sha:%s", sha)
	}

	if string(commitFmt) != "commit" {
		return fmt.Errorf("the git object does not have type commit for sha:%s type:%s", sha, string(commitFmt))
	}

	gitCommit, ok := commit.(*GitCommit)
	if !ok {
		return fmt.Errorf("expected *GitCommit type but got something else for sha:%s", sha)
	}

	shortHash := sha[:8]

	var messageBytes []byte
	for _, line := range (*(*gitCommit.Kvlm).Map)[""] {
		messageBytes = append(messageBytes, line...)
	}
	message := string(messageBytes)

	message = strings.ReplaceAll(message, "\\", "\\\\")
	message = strings.ReplaceAll(message, "\"", "\\\"")

	if newlineIndex := strings.Index(message, "\n"); newlineIndex != -1 {
		message = message[:newlineIndex]
	}

	fmt.Printf("  c_%s [label=\"%s: %s\"]\n", sha, shortHash, message)

	parents, ok := (*(*gitCommit.Kvlm).Map)["parent"]
	if !ok {
		return nil
	}

	for _, p := range parents {
		parentHash := string(p)
		fmt.Printf("  c_%s -> c_%s;\n", sha, parentHash)

		if err := LogGraphviz(repo, parentHash, seen); err != nil {
			return fmt.Errorf("error in logGraphviz for parent %s: %w", parentHash, err)
		}
	}

	return nil
}

func (repo *Repository) Rm(paths []string, withDelete bool, skipMissing bool) error {
	/*
		withDelete: default val is true
		skipMissing: default val is false
	*/
	index, err := repo.IndexRead()
	if err != nil {
		return err
	}

	workTree := repo.Worktree + string(filepath.Separator)
	absPaths := []string{}

	for _, path := range paths {
		pathAbs, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		if strings.HasPrefix(pathAbs, workTree) {
			absPaths = append(absPaths, pathAbs)
		} else {
			return fmt.Errorf("Cannot remove paths outside of worktree: %v\n", paths)
		}
	}

	keep := []GitIndexEntry{}
	remove := []string{}

	for _, e := range index.Entries {
		eFull := filepath.Join(workTree, e.Name)

		if slices.Contains(absPaths, eFull) {
			remove = append(remove, eFull)
			i := slices.Index(absPaths, eFull)
			slices.Delete(absPaths, i, i+1)
		} else {
			keep = append(keep, e)
		}
	}

	if len(absPaths) > 0 && !skipMissing {
		return fmt.Errorf("Cannot remove paths not in the index: %v", absPaths)
	}

	if withDelete {
		for _, path := range remove {
			err = os.Remove(path)
			if err != nil {
				return err
			}
		}
	}

	index.Entries = keep
	err = repo.IndexWrite(index)
	if err != nil {
		return err
	}

	return nil
}

type AddPathPairs struct {
	Abs string
	Rel string
}

func (repo *Repository) Add(paths []string) error {
	err := repo.Rm(paths, false, true)
	if err != nil {
		return err
	}

	workTree := repo.Worktree + string(filepath.Separator)
	pairs := []AddPathPairs{}

	for _, path := range paths {
		abs, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		isFile, err := utils.IsFile(abs)
		if err != nil {
			return err
		}

		if !(strings.HasPrefix(abs, workTree) || isFile) {
			return fmt.Errorf("Not a file, or outside the worktree: %v\n", paths)
		}
		rel, err := filepath.Rel(abs, repo.Worktree)
		if err != nil {
			return err
		}

		pairs = append(pairs, AddPathPairs{Abs: abs, Rel: rel})
	}

	index, err := repo.IndexRead()
	if err != nil {
		return err
	}

	for _, pair := range pairs {
		file, err := os.Open(pair.Abs)
		if err != nil {
			return err
		}
		defer file.Close()

		sha, err := ObjectHash(file, "blob", repo)
		if err != nil {
			return err
		}
		stat, err := os.Stat(pair.Abs)
		if err != nil {
			return err
		}

		sysStat := stat.Sys().(*syscall.Stat_t)

		ctime := fTime{
			Seconds:     uint32(sysStat.Ctim.Sec),
			Nanoseconds: uint32(sysStat.Ctim.Nsec % int64(1e9)),
		}
		mtime := fTime{
			Seconds:     uint32(sysStat.Mtim.Sec),
			Nanoseconds: uint32(sysStat.Mtim.Nsec % int64(1e9)),
		}

		entry := GitIndexEntry{
			CTime:            ctime,
			MTime:            mtime,
			Dev:              uint32(sysStat.Dev),
			Ino:              uint32(sysStat.Ino),
			ModeType:         0b1000,
			ModePerms:        0o644,
			UId:              sysStat.Uid,
			GId:              sysStat.Gid,
			FSize:            uint32(stat.Size()),
			Sha:              sha,
			FlagAssumedValid: false,
			FlagStage:        0,
			Name:             pair.Rel,
		}

		index.Entries = append(index.Entries, entry)
	}

	err = repo.IndexWrite(index)
	if err != nil {
		return err
	}

	return nil
}

func GitConfigRead() (*ini.File, error) {
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		xdgConfigHome = filepath.Join(os.Getenv("HOME"), ".config")
	}

	configFiles := []interface{}{
		filepath.Join(xdgConfigHome, "git", "config"),
		filepath.Join(os.Getenv("HOME"), ".gitconfig"),
	}

	conf, err := ini.LoadSources(ini.LoadOptions{Loose: true}, nil, configFiles...)
	if err != nil {
		return &ini.File{}, fmt.Errorf("failed to load config files: %w", err)
	}

	return conf, nil
}

func GitConfigUserGet(conf *ini.File) string {
	section, err := conf.GetSection("user")
	if err != nil {
		return ""
	}

	name := section.Key("name").String()
	email := section.Key("email").String()

	if name != "" && email != "" {
		return fmt.Sprintf("%s <%s>", name, email)
	}

	return ""
}

func (repo *Repository) TreeFromIndex(index *GitIndex) (string, error) {
	if index == nil {
		return "", fmt.Errorf("index is nil\n")
	}
	var err error
	contents := map[string][]GitIndexEntry{}
	contents[""] = []GitIndexEntry{}

	for _, e := range index.Entries {
		dirName := filepath.Dir(e.Name)

		key := dirName
		for key != "" && key != "." {
			if _, ok := contents[key]; !ok {
				contents[key] = []GitIndexEntry{}
			}
			key = filepath.Dir(key)
		}

		contents[dirName] = append(contents[dirName], e)
	}

	sortedPaths := make([]string, len(contents))
	for key := range contents {
		sortedPaths = append(sortedPaths, key)
	}
	slices.SortFunc(sortedPaths, func(a, b string) int {
		lenA := len(a)
		lenB := len(b)
		if lenA < lenB {
			return 1
		} else if lenA > lenB {
			return -1
		}
		return 0
	})

	sha := ""

	for _, path := range sortedPaths {
		tree := &GitTree{
			Items: []*GitTreeLeaf{},
		}

		for _, e := range contents[path] {
			var leaf *GitTreeLeaf
			if true {
				leafMode := fmt.Sprintf("%02o%04o", e.ModeType, e.ModePerms)
				leaf = &GitTreeLeaf{
					Mode: []byte(leafMode),
					Path: filepath.Base(e.Name),
					Sha:  e.Sha,
				}
			} else {
				leaf = &GitTreeLeaf{
					Mode: []byte("040000"),
					Path: filepath.Base(e.Name),
					Sha:  e.Sha,
				}
			}

			tree.Items = append(tree.Items, leaf)
		}

		sha, err = ObjectWrite(tree, repo)
		if err != nil {
			return "", err
		}

		parent := filepath.Dir(path)
		base := filepath.Base(path)
		_ = base

		contents[parent] = append(contents[parent], GitIndexEntry{})
	}

	return sha, nil
}

func (repo *Repository) CommitCreate(tree, parent, author, message string, timestamp time.Time) (string, error) {
	commit := &GitCommit{
		Kvlm: sharedtypes.NewKvlm(),
	}

	(*(*commit.Kvlm).Map)["tree"] = [][]byte{[]byte(tree)}
	if parent != "" {
		(*(*commit.Kvlm).Map)["parent"] = [][]byte{[]byte(parent)}
	}

	timestamp = time.Now()

	epochTime := timestamp.Unix()

	_, offset := timestamp.In(time.Local).Zone()

	hours := offset / 3600
	minutes := (offset % 3600) / 60

	var tz string
	if offset > 0 {
		tz = fmt.Sprintf("+%02d%02d", hours, minutes)
	} else {
		tz = fmt.Sprintf("-%02d%02d", -hours, -minutes)
	}

	author = fmt.Sprintf("%s %d %s", author, epochTime, tz)

	(*(*commit.Kvlm).Map)["author"] = [][]byte{[]byte(author)}
	(*(*commit.Kvlm).Map)["committer"] = [][]byte{[]byte(author)}
	(*(*commit.Kvlm).Map)[""] = [][]byte{[]byte(message)}

	return ObjectWrite(commit, repo)
}
