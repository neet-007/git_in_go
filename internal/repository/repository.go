package repository

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/neet-007/git_in_go/internal/sharedtypes"
	"github.com/neet-007/git_in_go/internal/utils"
	"gopkg.in/ini.v1"
)

type Repository struct {
	Worktree string
	Gitdir   string
	Conf     *ini.File
}

func NewRepository(path string, force bool) (*Repository, error) {
	repo := Repository{
		Worktree: path,
		Gitdir:   filepath.Join(path, ".git"),
	}

	info, err := os.Stat(repo.Gitdir)
	if err != nil && !force {
		return nil, err
	}

	if !(force || info.IsDir()) {
		return nil, fmt.Errorf("Not a Git repository %s", path)
	}

	cf := repo.RepoPath("config")

	cfg, err := ini.Load(cf)
	if err != nil {
		info, err = os.Stat(cf)
		if err != nil && !force {
			return nil, fmt.Errorf("Configuration file missing")
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

	return &repo, nil
}

func (repo *Repository) RepoPath(path ...string) string {
	return filepath.Join(append([]string{repo.Gitdir}, path...)...)
}

func (repo *Repository) RepoFile(mkdir bool, path ...string) (string, error) {
	/*
		mkdir: default val is false
	*/

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

func CreateRepo(path string) (*Repository, error) {
	repo, err := NewRepository(path, true)

	if err != nil {
		return nil, err
	}

	info, err := os.Stat(repo.Worktree)
	if err != nil {
		os.MkdirAll(path, 0755)
		info, err = os.Stat(repo.Worktree)

		if err != nil {
			return nil, fmt.Errorf("error even after making file %w", err)
		}
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("Not a directory %s", path)
	}

	_, err = os.Stat(repo.Gitdir)
	if err == nil {
		_, err := os.ReadDir(repo.Gitdir)

		if err == nil {
			return nil, fmt.Errorf("%s is not empty", path)
		}
	}

	os.MkdirAll(repo.Gitdir, 0755)

	_, err = repo.RepoDir(true, "branches")
	if err != nil {
		return nil, err
	}

	_, err = repo.RepoDir(true, "objects")
	if err != nil {
		return nil, err
	}

	_, err = repo.RepoDir(true, "refs", "tags")
	if err != nil {
		return nil, err
	}

	_, err = repo.RepoDir(true, "refs", "heads")
	if err != nil {
		return nil, err
	}

	dir := repo.RepoPath("description")

	file, err := os.Create(dir)
	if err != nil {
		return nil, fmt.Errorf("Failed to create file: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteString("Unnamed repository; edit this file 'description' to name the repository.\n"); err != nil {
		return nil, fmt.Errorf("Failed to write to file: %v", err)
	}

	dir = repo.RepoPath("HEAD")

	file, err = os.Create(dir)
	if err != nil {
		return nil, fmt.Errorf("Failed to create file: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteString("ref: refs/heads/master\n"); err != nil {
		return nil, fmt.Errorf("Failed to write to file: %v", err)
	}

	dir = repo.RepoPath("config")

	config, err := repoDefaultConfig()
	if err != nil {
		return nil, err
	}

	err = config.SaveTo(dir)
	if err != nil {
		return nil, err
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

func (repo *Repository) CatFile(objName string, fmtType string) {
	objSha, err := repo.ObjectFind(objName, fmtType, true)
	if err != nil {
		log.Fatalf("Error while cat-file read: %v\n", err)
	}

	obj, err := repo.ObjectRead(objSha)
	if err != nil || obj == nil {
		log.Fatalf("Error while cat-file read: %v\n", err)
	}

	data, err := obj.Serialize()
	if err != nil || obj == nil {
		log.Fatalf("Error while cat-file serialize: %v\n", err)
	}

	fmt.Printf("%v\n", string(data))
}

func (repo *Repository) LsTree(path string, recursive bool, prefix string) error {
	/*
		recursive: default val is false
		prefix: default val is ""
	*/

	sha, err := repo.ObjectFind(path, "tree", true)
	if err != nil {
		return err
	}

	obj, err := repo.ObjectRead(sha)

	if err != nil {
		return err
	}

	if objFmt, err := obj.GetFmt(); err != nil || string(objFmt) != "tree" {
		if err != nil {
			return fmt.Errorf("Error while ls-tree err:%w\n", err)
		}
		return fmt.Errorf("Error while ls-tree expected object to be tree got:%s\n", objFmt)
	}

	objTree, ok := obj.(*GitTree)
	if !ok {
		return fmt.Errorf("Error while ls-tree could not cast as tree\n")
	}

	for _, item := range objTree.Items {
		var objType []byte
		var objTypeStr string

		if len(item.Mode) == 5 {
			objType = item.Mode[:1]
		} else {
			objType = item.Mode[:2]
		}

		objTypeNumStr := string(objType)
		switch {
		case objTypeNumStr == "04":
			objTypeStr = "tree"
		case objTypeNumStr == "10":
			objTypeStr = "blob"
		case objTypeNumStr == "12":
			objTypeStr = "blob"
		case objTypeNumStr == "16":
			objTypeStr = "commit"
		default:
			return fmt.Errorf("unknow type %s bytes %v  for path %s\n", objTypeNumStr, objType, path)
		}

		if !(recursive && objTypeStr == "tree") {
			fmt.Printf("%s %s %s\t %s\n", strings.Repeat("0", (6-len(item.Mode)))+string(item.Mode), objTypeStr, item.Sha, filepath.Join(prefix, item.Path))
			continue
		}

		err := repo.LsTree(item.Sha, recursive, filepath.Join(prefix, item.Path))
		if err != nil {
			return err
		}
	}

	return nil
}

func (repo *Repository) TreeCheckout(tree *GitTree, path string) error {
	if tree == nil {
		return fmt.Errorf("tree is nil at path:%s\n", path)
	}

	for _, item := range tree.Items {
		if item == nil {
			return fmt.Errorf("tree item is nil at path:%s\n", path)
		}

		obj, err := repo.ObjectRead(item.Sha)
		if err != nil {
			return err
		}

		dest := filepath.Join(path, item.Path)

		objFmt, err := obj.GetFmt()
		if err != nil {
			return err
		}

		if string(objFmt) == "tree" {
			if err = os.MkdirAll(dest, 0755); err != nil {
				return err
			}
			objTree := obj.(*GitTree)
			repo.TreeCheckout(objTree, dest)
			continue
		}

		if string(objFmt) == "blob" {
			//IMPORTANT ADD SYM LINK
			file, err := os.Create(dest)
			if err != nil {
				return err
			}
			defer file.Close()

			objBlob := obj.(*GitBlob)
			_, err = file.Write(objBlob.BlobData)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type RefRes struct {
	Name string
	Dir  *map[string]RefRes
}

func (repo *Repository) RefResolve(path string) (string, error) {
	var err error
	if !filepath.IsAbs(path) {
		path, err = repo.RepoFile(false, path)
		if err != nil {
			return "", err
		}
	}
	path = strings.TrimSpace(path)

	ok, err := utils.IsFile(path)
	if err != nil {
		return "", err
	}

	if !ok {
		return "", fmt.Errorf("not a file path:%s\n", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(string(data), "ref: ") {
		return repo.RefResolve(strings.Replace(string(data), "ref: ", "", 1))
	}

	return string(data[:len(data)-1]), nil
}

func (repo *Repository) RefList(path string) (*map[string]RefRes, error) {
	/*
		path: default value is ""
	*/
	var err error
	if path == "" {
		path, err = repo.RepoDir(false, "refs")
		if err != nil {
			return &map[string]RefRes{}, err
		}
	}

	ret := map[string]RefRes{}

	entries, err := os.ReadDir(path)
	if err != nil {
		return &map[string]RefRes{}, err
	}

	for _, e := range entries {
		can := filepath.Join(path, e.Name())
		info, err := os.Stat(can)
		if err != nil {
			return &map[string]RefRes{}, err
		}

		if info.IsDir() {
			dir, err := repo.RefList(can)
			if err != nil {
				return &map[string]RefRes{}, err
			}

			ret[e.Name()] = RefRes{Name: "", Dir: dir}
			continue
		}

		name, err := repo.RefResolve(can)
		if err != nil {
			return &map[string]RefRes{}, err
		}

		ret[e.Name()] = RefRes{Name: name, Dir: nil}
	}

	return &ret, nil
}

func (repo *Repository) ShowRefs(refs *map[string]RefRes, withHash bool, prefix string) error {
	/*
		withHash: defaul true
		prefix: defalut ""
	*/
	if refs == nil {
		return fmt.Errorf("ref is nil")
	}

	for key, val := range *refs {
		if val.Dir != nil {
			err := repo.ShowRefs(val.Dir, withHash, prefix)
			if err != nil {
				return err
			}
			continue
		}
		hash := ""
		if withHash {
			hash = val.Name
		}
		prefix_ := ""
		if prefix != "" {
			prefix_ = prefix + "/"
		}
		fmt.Printf("%s %s %s\n", hash, prefix_, key)
	}

	return nil
}

func (repo *Repository) TagCreate(name string, ref string, createTagObject bool) error {
	/*
		createTagObject: default value is false
	*/
	sha, err := repo.ObjectFind(ref, "", false)
	if err != nil {
		return err
	}

	if createTagObject {
		tag := &GitTag{}
		tag.Kvlm = sharedtypes.NewKvlm()
		tag.Kvlm.InsertAndSort("object", [][]byte{[]byte(sha)})
		tag.Kvlm.InsertAndSort("type", [][]byte{[]byte("commit")})
		tag.Kvlm.InsertAndSort("tag", [][]byte{[]byte(name)})
		tag.Kvlm.InsertAndSort("tagger", [][]byte{[]byte("test test@example.com")})
		tag.Kvlm.InsertAndSort("", [][]byte{[]byte("A tag generated by wyag, which won't let you customize the message!")})

		tagSha, err := ObjectWrite(tag, repo)
		if err != nil {
			return err
		}

		err = repo.RefCreate("tags/"+name, tagSha)
		if err != nil {
			return err
		}
		return nil
	}

	err = repo.RefCreate("tags/"+name, sha)
	if err != nil {
		return err
	}

	return nil
}

func (repo *Repository) RefCreate(refName string, sha string) error {
	fileName, err := repo.RepoFile(true, "refs/"+refName)
	if err != nil {
		return err
	}

	file, err := os.Create(fileName)
	defer file.Close()

	file.Write([]byte(sha + "\n"))
	return nil
}

func (repo *Repository) ObjectResolve(name string) ([]string, error) {
	if strings.TrimSpace(name) == "" {
		return []string{}, fmt.Errorf("name is empty\n")
	}

	candidates := []string{}

	if name == "HEAD" {
		res, err := repo.RefResolve(name)
		if err != nil {
			return []string{}, err
		}

		return []string{res}, nil
	}

	lenName := len(name)
	if lenName < 4 || lenName > 40 {
		return []string{}, fmt.Errorf("len name is not valid must be 4 < %d < 40 name:%s\n", lenName, name)
	}

	_, err := hex.DecodeString(name)
	if err == nil {
		name = strings.ToLower(name)
		prefix := name[:2]
		path, err := repo.RepoDir(false, "objects", prefix)
		if err != nil {
			return []string{}, err
		}

		rem := name[2:]
		entries, err := os.ReadDir(path)
		if err != nil {
			return []string{}, err
		}

		for _, e := range entries {
			if strings.HasPrefix(e.Name(), rem) {
				candidates = append(candidates, prefix+e.Name())
			}
		}
	}

	asTag, err := repo.RefResolve("refs/tags/" + name)
	if err == nil {
		candidates = append(candidates, asTag)
	}

	asBranch, err := repo.RefResolve("refs/heads/" + name)
	if err == nil {
		candidates = append(candidates, asBranch)
	}

	return candidates, nil
}

func FindRepo(path string, required bool) (*Repository, error) {
	/*
		path: default val is .
		required default val is true
	*/

	path, err := utils.RealPath(path)

	if err != nil {
		return nil, err
	}

	info, err := os.Stat(filepath.Join(path, ".git"))
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		repo, err := NewRepository(path, false)
		if err != nil {
			return nil, err
		}

		return repo, nil
	}

	parent := filepath.Clean(filepath.Join(path, ".."))
	if parent == path {
		if required {
			return nil, fmt.Errorf("No git directory")
		}

		return nil, fmt.Errorf("No git direcotory but not required")
	}

	return FindRepo(path, required)
}
