package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

type Repository struct {
	Worktree string
	Gitdir   string
	Conf     *ini.File
}

func NewRepository(path string, force bool) (Repository, error) {
	repo := Repository{
		Worktree: path,
		Gitdir:   filepath.Join(path, ".git"),
	}

	info, err := os.Stat(repo.Gitdir)
	if err != nil && !force {
		return Repository{}, err
	}

	if !(force || info.IsDir()) {
		return Repository{}, fmt.Errorf("Not a Git repository %s", path)
	}

	cf := repo.RepoPath("config")

	cfg, err := ini.Load(cf)
	if err != nil {
		info, err = os.Stat(cf)
		if err != nil && !force {
			return Repository{}, fmt.Errorf("Configuration file missing")
		}
	}

	repo.Conf = cfg

	if !force {
		vers, err := repo.Conf.Section("core").Key("repositoryformatversion").Int()
		if err != nil {
			panic("repositoryformatversion not found or invalid")
		}
		if vers != 0 {
			panic(fmt.Sprintf("Unsupported repositoryformatversion %d", vers))
		}
	}

	return repo, nil
}

func (repo *Repository) RepoPath(path ...string) string {
	return filepath.Join(append([]string{repo.Gitdir}, path...)...)
}

func (repo *Repository) RepoFile(mkdir bool, path ...string) (string, error) {

	if _, err := repo.RepoDir(mkdir, path[:len(path)-1]...); err != nil {
		return "", err
	}

	return repo.RepoPath(path...), nil
}

func (repo *Repository) RepoDir(mkdir bool, path ...string) (string, error) {
	pathLocal := repo.RepoPath(path...)

	info, err := os.Stat(pathLocal)
	if err != nil {
		if !mkdir {
			return "", fmt.Errorf("dir does not exist and mkdir is false")
		}

		os.Mkdir(pathLocal, 0755)
		return pathLocal, nil
	}

	if !info.IsDir() {
		return "", fmt.Errorf("Not a directory %s", pathLocal)
	}

	return pathLocal, nil
}

func CreateRepo(path string) (Repository, error) {
	repo, err := NewRepository(path, true)

	if err != nil {
		return Repository{}, err
	}

	info, err := os.Stat(repo.Worktree)
	if err != nil {
		os.MkdirAll(path, 0755)
		info, err = os.Stat(repo.Worktree)

		if err != nil {
			return Repository{}, fmt.Errorf("error even after making file %w", err)
		}
	}

	if !info.IsDir() {
		return Repository{}, fmt.Errorf("Not a directory %s", path)
	}

	_, err = os.Stat(repo.Gitdir)
	if err == nil {
		_, err := os.ReadDir(repo.Gitdir)

		if err == nil {
			return Repository{}, fmt.Errorf("%s is not empty", path)
		}
	}

	os.MkdirAll(repo.Gitdir, 0755)

	_, err = repo.RepoDir(true, "branches")
	if err != nil {
		return Repository{}, err
	}

	_, err = repo.RepoDir(true, "objects")
	if err != nil {
		return Repository{}, err
	}

	_, err = repo.RepoDir(true, "refs", "tags")
	if err != nil {
		return Repository{}, err
	}

	_, err = repo.RepoDir(true, "refs", "heads")
	if err != nil {
		return Repository{}, err
	}

	dir := repo.RepoPath("description")

	file, err := os.Create(dir)
	if err != nil {
		return Repository{}, fmt.Errorf("Failed to create file: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteString("Unnamed repository; edit this file 'description' to name the repository.\n"); err != nil {
		return Repository{}, fmt.Errorf("Failed to write to file: %v", err)
	}

	dir = repo.RepoPath("HEAD")

	file, err = os.Create(dir)
	if err != nil {
		return Repository{}, fmt.Errorf("Failed to create file: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteString("ref: refs/heads/master\n"); err != nil {
		return Repository{}, fmt.Errorf("Failed to write to file: %v", err)
	}

	dir = repo.RepoPath("config")

	config, err := repoDefaultConfig()
	if err != nil {
		return Repository{}, err
	}

	err = config.SaveTo(dir)
	if err != nil {
		return Repository{}, err
	}

	return repo, nil
}

func repoDefaultConfig() (*ini.File, error) {
	cfg := ini.Empty()

	coreSection, err := cfg.NewSection("core")
	if err != nil {
		return ini.Empty(), err
	}

	coreSection.Key("repositoryformatversion").SetValue("0")
	coreSection.Key("filemode").SetValue("false")
	coreSection.Key("bare").SetValue("false")

	return cfg, nil
}
