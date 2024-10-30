package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/neet-007/git_in_go/internal/bridges"
)

func main() {

	args := os.Args

	switch args[1] {
	case "add":
		bridges.Cmd_add(args[0])
	case "cat-file":
		bridges.CmdCatFile(args...)
	case "check-ignore":
		bridges.Cmd_check_ignore(args[0])
	case "checkout":
		bridges.Cmd_checkout(args[0])
	case "commit":
		bridges.Cmd_commit(args[0])
	case "hash-object":
		var writeFlag bool
		var typeFlag string

		hashObjectCmd := flag.NewFlagSet("hash-object", flag.ExitOnError)
		hashObjectCmd.BoolVar(&writeFlag, "w", false, "Write the hash to the repo")
		hashObjectCmd.StringVar(&typeFlag, "t", "", "Specify the type of the hash")

		hashObjectCmd.Parse(args[2:])

		positionalArgs := hashObjectCmd.Args()

		if len(positionalArgs) == 0 {
			fmt.Println("You must provide a file path for hash-object.")
			os.Exit(1)
		}

		bridges.CmdHashObject(writeFlag, typeFlag, positionalArgs[0])
	case "init":
		bridges.CmdInit(args...)
	case "log":
		bridges.Cmd_log(args[0])
	case "ls-files":
		bridges.Cmd_ls_files(args[0])
	case "ls-tree":
		bridges.Cmd_ls_tree(args[0])
	case "rev-parse":
		bridges.Cmd_rev_parse(args[0])
	case "rm":
		bridges.Cmd_rm(args[0])
	case "show-ref":
		bridges.Cmd_show_ref(args[0])
	case "status":
		bridges.Cmd_status(args[0])
	case "tag":
		bridges.Cmd_tag(args[0])
	default:
		log.Fatal("unkown command")
	}
}
