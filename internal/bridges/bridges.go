package bridges

import (
	"fmt"

	"github.com/neet-007/git_in_go/internal/repository"
)

func Cmd_add(args string) {

}
func Cmd_cat_file(args string) {

}
func Cmd_check_ignore(args string) {

}
func Cmd_checkout(args string) {

}
func Cmd_commit(args string) {

}
func Cmd_hash_object(args string) {

}
func Cmd_init(args ...string) {
	_, err := repository.CreateRepo(args[1])
	if err != nil {
		fmt.Printf("Error: %w\n", err)
		return
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
