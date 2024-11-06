package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"github.com/neet-007/git_in_go/internal/utils"
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
