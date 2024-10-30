package bridges

import (
	"fmt"
	"log"
	"os"

	"github.com/neet-007/git_in_go/internal/repository"
)

func Cmd_add(args string) {

}

func CmdCatFile(args ...string) {
	repo, err := repository.FindRepo(".", true)
	if err != nil {
		log.Fatalf("Error with cat-file: %v", err)
	}

	repo.CatFile(args[1], args[2])

}

func Cmd_check_ignore(args string) {

}
func Cmd_checkout(args string) {

}
func Cmd_commit(args string) {

}
func CmdHashObject(write bool, typeName string, path string) {
	var repo *repository.Repository
	var err error

	if write {
		repo, err = repository.FindRepo(".", true)
		if err != nil {
			log.Fatalf("Error with hash object: %v", err)
		}
	} else {
		repo = nil
	}

	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Error with hash object: %v", err)
	}

	defer file.Close()

	sha, err := repository.ObjectHash(file, typeName, repo)
	if err != nil {
		log.Fatalf("Error with hash object: %v", err)
	}

	fmt.Printf("%s\n", sha)
}
func CmdInit(args ...string) {
	_, err := repository.CreateRepo(args[1])
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	fmt.Println("empty repo is initinlized")
}
func Cmd_log(args string) {

}
func Cmd_ls_files(args string) {

}
func Cmd_ls_tree(args string) {

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
