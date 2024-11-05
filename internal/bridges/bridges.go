package bridges

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/neet-007/git_in_go/internal/repository"
)

func Cmd_add(args string) {

}

func CmdCatFile(args ...string) {
	repo, err := repository.FindRepo(".", true)
	if err != nil {
		log.Fatalf("Error with cat-file: %v\n", err)
	}

	repo.CatFile(args[3], args[2])

}

func CmdCheckIgnore(paths ...string) {
	repo, err := repository.FindRepo(".", true)
	if err != nil {
		log.Fatalf("Error while check-ignore: %v\n", err)
	}

	rules, err := repo.GitIgnoreRead()

	for _, path := range paths {
		res, err := repository.CheckIgnore(rules, path)
		if err != nil && !errors.Is(err, repository.GitIgnoreDefaultCheck) {
			log.Fatalf("Error while check-ignore: %v\n", err)
		}
		if err == nil && res {
			fmt.Printf("%s ", path)
		}
	}
}

func CmdCheckout(commit string, path string) {
	repo, err := repository.FindRepo(".", true)
	if err != nil {
		log.Fatalf("Error with checkout: %v\n", err)
	}

	objSha, err := repo.ObjectFind(commit, "", true)
	if err != nil {
		log.Fatalf("Error with checkout: %v\n", err)
	}

	obj, err := repo.ObjectRead(objSha)
	if err != nil {
		log.Fatalf("Error with checkout: %v\n", err)
	}

	objFmt, err := obj.GetFmt()
	if err != nil {
		log.Fatalf("Error with checkout: %v\n", err)
	}

	if string(objFmt) == "commit" {
		objTree := obj.(*repository.GitCommit)
		obj, err = repo.ObjectRead(string(bytes.Join((*(*objTree.Kvlm).Map)["tree"], nil)))
		if err != nil {
			log.Fatalf("Error with checkout: %v\n", err)
		}
	}

	if _, err := os.Stat(path); err == nil {
		info, err := os.Stat(path)
		if err != nil {
			log.Fatalf("Error with checkout: %v\n", err)
		}
		if !info.IsDir() {
			log.Fatalf("Error with checkout: dir does not exits path:%s\n", path)
		}

		contents, err := os.ReadDir(path)
		if err != nil {
			log.Fatalf("Error with checkout: %v\n", err)
		}
		if len(contents) > 0 {
			log.Fatalf("Error with checkout: dir is not empty path:%s\n", path)
		}
	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			log.Fatalf("Error with checkout: %v\n", err)
		}
	} else {
		log.Fatalf("Error with checkout: %v\n", err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Println("Error resolving absolute path:%s, err:%s\n", path, err)
		return
	}

	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		fmt.Println("Error resolving real path:\n", err)
		return
	}

	objTree := obj.(*repository.GitTree)
	err = repo.TreeCheckout(objTree, realPath)
	if err != nil {
		fmt.Println("Error resolving real path:\n", err)
	}
}

func Cmd_commit(args string) {

}

func CmdHashObject(write bool, typeName string, path string) {
	var repo *repository.Repository
	var err error

	if write {
		repo, err = repository.FindRepo(".", true)
		if err != nil {
			log.Fatalf("Error with hash object: %v\n", err)
		}
	} else {
		repo = nil
	}

	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Error with hash object: %v\n", err)
	}

	defer file.Close()

	sha, err := repository.ObjectHash(file, typeName, repo)
	if err != nil {
		log.Fatalf("Error with hash object: %v\n", err)
	}

	fmt.Printf("%s\n", sha)
}

func CmdInit(args ...string) {
	_, err := repository.CreateRepo(args[1])
	if err != nil {
		log.Fatalf("Error while initlizaing repo: %v\n", err)
	}

	fmt.Println("empty repo is initinlized\n")
}

func CmdLog(commit string) {
	repo, err := repository.FindRepo(".", true)
	if err != nil {
		log.Fatalf("Error while logging: %v\n", err)
	}

	fmt.Println("digraph gitlog{\n node[shape=rect]")
	path, err := repo.ObjectFind(commit, "", true)
	if err != nil {
		log.Fatalf("Error while logging: %v\n", err)
	}

	repository.LogGraphviz(repo, path, &map[string]byte{})
	fmt.Println("}")
}

func CmdLsFiles(verbose bool) {
	repo, err := repository.FindRepo(".", true)
	if err != nil {
		log.Fatalf("Error while ls-files: %v\n", err)
	}

	index, err := repo.IndexRead()
	if err != nil {
		log.Fatalf("Error while ls-files: %v\n", err)
	}
	if index == nil {
		log.Fatalf("Error while ls-files: index is nil\n")
	}

	if verbose {
		fmt.Printf("Index file format v%d, containing %d entries.\n", index.Version, len(index.Entries))
	}

	for _, e := range index.Entries {
		fmt.Printf("%s\n", e.Name)

		if verbose {
			//I DONT KNOW HOW THEY PRINT IT
		}
	}
}

func CmdLsTree(path string, recursive bool) {
	repo, err := repository.FindRepo(".", true)
	if err != nil {
		log.Fatalf("Error while ls-tree for path:%s err:%w\n", path, err)
	}

	err = repo.LsTree(path, recursive, "")
	if err != nil {
		log.Fatalf("Error while ls-tree for path:%s err:%w\n", path, err)
	}

}

func Cmd_rev_parse(typeArg string, name string) {
	repo, err := repository.FindRepo(".", true)
	if err != nil {
		log.Fatalf("Error while rev-parse name:%s, type:%s, %v", name, typeArg, err)
	}

	sha, err := repo.ObjectFind(name, typeArg, true)
	if err != nil {
		log.Fatalf("Error while rev-parse name:%s, type:%s, %v", name, typeArg, err)
	}

	fmt.Printf("%s\n", sha)
}

func Cmd_rm(args string) {

}

func Cmd_show_ref(args string) {
	repo, err := repository.FindRepo(".", true)
	if err != nil {
		log.Fatalf("Error while show-ref %v\n", err)
	}

	refs, err := repo.RefList("")
	if err != nil {
		log.Fatalf("Error while show-ref %v\n", err)
	}

	err = repo.ShowRefs(refs, true, "refs")
	if err != nil {
		log.Fatalf("Error while show-ref %v\n", err)
	}
}

func Cmd_status(args string) {

}

func Cmd_tag(name string, object string, createTagObject bool) {
	/*
		name: default val is ""
		object: default val is ""
		createTagObject: default val is false
	*/
	repo, err := repository.FindRepo(".", true)
	if err != nil {
		log.Fatalf("Error while tag %v\n", err)
	}

	if name != "" {
		err := repo.TagCreate(name, object, createTagObject)
		if err != nil {
			log.Fatalf("Error while tag %v\n", err)
		}

		return
	}

	refs, err := repo.RefList("")
	if err != nil {
		log.Fatalf("Error while tag %v\n", err)
	}

	if refs == nil {
		log.Fatalf("Error while tag refs is nil name:%s object:%s createTagObject:%v\n", name, object, createTagObject)
	}

	err = repo.ShowRefs((*refs)["tags"].Dir, false, "")
	if err != nil {
		log.Fatalf("Error while tag %v\n", err)
	}
}
