package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func (repo *Repository) GetActiveBranch() (string, error) {
	headFile, err := repo.RepoFile(false, "HEAD")
	if err != nil {
		return "", err
	}

	fileData, err := os.ReadFile(headFile)
	if err != nil {
		return "", err
	}

	data := string(fileData)
	if strings.HasPrefix(data, "ref: refs/heads/") {
		return data[16:], nil
	}

	return "", nil
}

func (repo *Repository) StatusBranch() error {
	branch, err := repo.GetActiveBranch()
	if err != nil {
		return err
	}

	if branch != "" {
		fmt.Printf("On branch %s.", branch)
	}

	sha, err := repo.ObjectFind("HEAD", "", true)
	if err != nil {
		return err
	}

	fmt.Printf("HEAD detached at %s.\n", sha)
	return nil
}

func (repo *Repository) treeToDict(ref, prefix string) (*map[string]string, error) {
	/*
		prefix: default val is ""
	*/
	ret := &map[string]string{}
	sha, err := repo.ObjectFind(ref, "tree", true)
	if err != nil {
		return &map[string]string{}, err
	}

	obj, err := repo.ObjectRead(sha)
	if err != nil {
		return &map[string]string{}, err
	}

	tree, ok := obj.(*GitTree)
	if !ok {
		return &map[string]string{}, fmt.Errorf("could not read obj as tree for sha:%s ref:%s prefix:%s\n", sha, ref, prefix)
	}

	for _, leaf := range tree.Items {
		fullPath := filepath.Join(prefix, leaf.Path)

		if string(leaf.Mode[:2]) == "04" {
			subtreeDict, err := repo.treeToDict(leaf.Sha, fullPath)
			if err != nil {
				return &map[string]string{}, err
			}

			for k, v := range *subtreeDict {
				(*ret)[k] = v
			}
		} else {
			(*ret)[fullPath] = leaf.Sha
		}
	}

	return ret, nil
}

func (repo *Repository) StatusHeadIndex(index *GitIndex) error {
	fmt.Println("Changes to be committed:")

	head, err := repo.treeToDict("HEAD", "")
	if err != nil {
		return err
	}

	for _, e := range index.Entries {
		if _, ok := (*head)[e.Name]; ok {
			if (*head)[e.Name] != e.Sha {
				fmt.Printf("  modified:%s\n", e.Name)
			}
			delete(*head, e.Name)
		} else {
			fmt.Printf("  added:  %s\n", e.Name)
		}
	}

	for e := range *head {
		fmt.Printf("  deleted: %s\n", e)
	}

	return nil
}

func (repo *Repository) StatusIndexWorktree(index *GitIndex) error {
	fmt.Println("Changes not staged for commit:")

	ignore, err := repo.GitIgnoreRead()
	if err != nil {
		return err
	}

	gitDirPrefix := repo.Gitdir + string(os.PathSeparator)
	allFiles := []string{}

	err = filepath.Walk(repo.Worktree, func(fullPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fullPath == repo.Gitdir || strings.HasPrefix(fullPath, gitDirPrefix) {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(repo.Worktree, fullPath)
		if err != nil {
			return err
		}

		allFiles = append(allFiles, relPath)
		return nil
	})

	if err != nil {
		return err
	}

	for _, e := range index.Entries {
		fullPath := filepath.Join(repo.Worktree, e.Name)

		stat, err := os.Stat(fullPath)
		if err != nil {
			fmt.Printf("deleted: %s\n", e.Name)
			continue
		}

		ctimeNs := e.CTime.Seconds*1_000_000_000 + e.CTime.Nanoseconds
		mtimeNs := e.MTime.Seconds*1_000_000_000 + e.MTime.Nanoseconds

		if stat.ModTime().UnixNano() != int64(ctimeNs) || stat.ModTime().UnixNano() != int64(mtimeNs) {
			//@FIXME This *will* crash on symlinks to dir.
			file, err := os.Open(fullPath)
			if err != nil {
				return err
			}
			defer file.Close()

			newSha, err := ObjectHash(file, "blob", nil)
			if err != nil {
				return err
			}

			if e.Sha != newSha {
				fmt.Printf("  modified:%s\n", e.Name)
			}
		}

		i := slices.Index(allFiles, e.Name)
		if i != -1 {
			slices.Delete(allFiles, i, i+1)
		}

	}

	fmt.Println()
	fmt.Println("Untracked files:")

	for _, f := range allFiles {
		// @TODO If a full directory is untracked, we should display its name without its contents.
		res, err := CheckIgnore(ignore, f)
		if err == nil && res {
			fmt.Printf(" %s\n", f)
		}
	}

	return nil
}
