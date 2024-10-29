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
		log.Fatalf("Error with cat-file: %w", err)
	}

	repo.CatFile(args[1], args[2])

}

func Cmd_check_ignore(args string) {

}
func Cmd_checkout(args string) {

}
func Cmd_commit(args string) {

}
func CmdHashObject(args ...string) {
	var repo *repository.Repository
	var err error

	if len(args) > 1 && args[1] != "" {
		repo, err = repository.FindRepo(".", true)
		if err != nil {
			log.Fatalf("Error with hash object: %v", err)
		}
	} else {
		repo = nil
	}

	file, err := os.Open(args[2])
	if err != nil {
		log.Fatalf("Error with hash object: %w", err)
	}

	defer file.Close()

	sha, err := repository.ObjectHash(file, args[3], repo)
	if err != nil {
		log.Fatalf("Error with hash object: %w", err)
	}

	fmt.Printf("%s", sha)
}
func CmdInit(args ...string) {
	_, err := repository.CreateRepo(args[1])
	if err != nil {
		log.Fatalf("Error: %w\n", err)
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
