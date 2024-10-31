package bridges

import (
	"bytes"
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
		log.Fatalf("Error with cat-file: %v", err)
	}

	repo.CatFile(args[3], args[2])

}

func Cmd_check_ignore(args string) {

}

func CmdCheckout(commit string, path string) {
	repo, err := repository.FindRepo(".", true)
	if err != nil {
		log.Fatalf("Error with checkout: %v", err)
	}

	obj, err := repo.ObjectRead(repo.ObjectFind(commit, "", false))
	if err != nil {
		log.Fatalf("Error with checkout: %v", err)
	}

	objFmt, err := obj.GetFmt()
	if err != nil {
		log.Fatalf("Error with checkout: %v", err)
	}

	if string(objFmt) == "commit" {
		objTree := obj.(*repository.GitCommit)
		obj, err = repo.ObjectRead(string(bytes.Join((*objTree.Kvlm)["tree"], nil)))
		if err != nil {
			log.Fatalf("Error with checkout: %v", err)
		}
	}

	if _, err := os.Stat(path); err == nil {
		info, err := os.Stat(path)
		if err != nil {
			log.Fatalf("Error with checkout: %v", err)
		}
		if !info.IsDir() {
			log.Fatalf("Error with checkout: %v", err)
		}

		contents, err := os.ReadDir(path)
		if err != nil {
			log.Fatalf("Error with checkout: %v", err)
		}
		if len(contents) > 0 {
			log.Fatalf("Error with checkout: %v", err)
		}
	} else if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			log.Fatalf("Error with checkout: %v", err)
		}
	} else {
		log.Fatalf("Error with checkout: %v", err)
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		fmt.Println("Error resolving absolute path:", err)
		return
	}

	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		fmt.Println("Error resolving real path:", err)
		return
	}

	objTree := obj.(*repository.GitTree)
	err = repo.TreeCheckout(objTree, realPath)
	if err != nil {
		fmt.Println("Error resolving real path:", err)
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
	repository.LogGraphviz(repo, commit, &map[string]byte{})
	fmt.Println("}")
}

func Cmd_ls_files(args string) {

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

func Cmd_rev_parse(args string) {

}

func Cmd_rm(args string) {

}

func Cmd_show_ref(args string) {

}

func Cmd_status(args string) {

}

func Cmd_tag(args string) {

}
