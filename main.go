package main

import (
	"flag"
	"log"
	"os"

	"github.com/neet-007/git_in_go/internal/bridges"
)

func main() {

	args := os.Args

	switch args[1] {
	case "add":
		bridges.CmdAdd(args[2:]...)
	case "cat-file":
		bridges.CmdCatFile(args...)
	case "check-ignore":
		bridges.CmdCheckIgnore(args[2:]...)
	case "checkout":
		bridges.CmdCheckout(args[2], args[3])
	case "commit":
		var messageFlag string

		commitCmd := flag.NewFlagSet("hash-object", flag.ExitOnError)
		commitCmd.StringVar(&messageFlag, "m", "", "the message of the commit")

		commitCmd.Parse(args[2:])

		bridges.CmdCommit(messageFlag)
	case "hash-object":
		var writeFlag bool
		var typeFlag string

		hashObjectCmd := flag.NewFlagSet("hash-object", flag.ExitOnError)
		hashObjectCmd.BoolVar(&writeFlag, "w", false, "Write the hash to the repo")
		hashObjectCmd.StringVar(&typeFlag, "t", "", "Specify the type of the hash")

		hashObjectCmd.Parse(args[2:])

		positionalArgs := hashObjectCmd.Args()

		if len(positionalArgs) == 0 {
			log.Fatal("You must provide a file path for hash-object.")
		}

		bridges.CmdHashObject(writeFlag, typeFlag, positionalArgs[0])
	case "init":
		bridges.CmdInit(args...)
	case "log":
		if len(args) < 3 {
			bridges.CmdLog("HEAD")
		} else {
			bridges.CmdLog(args[2])
		}
	case "ls-files":
		var verboseFlag bool

		lsFilesCmd := flag.NewFlagSet("ls-files", flag.ExitOnError)
		lsFilesCmd.BoolVar(&verboseFlag, "verbose", false, "recursively print the tree")

		lsFilesCmd.Parse(args[2:])

		positionalArgs := lsFilesCmd.Args()

		if len(positionalArgs) != 0 {
			log.Fatal("You must not provide positional args")
		}

		bridges.CmdLsFiles(verboseFlag)
	case "ls-tree":
		var recursiceFlag bool

		lsTreeCmd := flag.NewFlagSet("ls-tree", flag.ExitOnError)
		lsTreeCmd.BoolVar(&recursiceFlag, "r", false, "recursively print the tree")

		lsTreeCmd.Parse(args[2:])

		positionalArgs := lsTreeCmd.Args()

		if len(positionalArgs) == 0 {
			log.Fatal("You must provide a dir or file path for ls-tree")
		}

		bridges.CmdLsTree(positionalArgs[0], recursiceFlag)
	case "rev-parse":
		var typeFlag string

		revParseCmd := flag.NewFlagSet("rev-parse", flag.ExitOnError)
		revParseCmd.StringVar(&typeFlag, "git-type", "", "expceted type")

		revParseCmd.Parse(args[2:])

		positionalArgs := revParseCmd.Args()

		if len(positionalArgs) == 0 {
			log.Fatal("You must provide a dir or file path for rev-parse")
		}

		bridges.CmdRevParse(typeFlag, positionalArgs[0])
	case "rm":
		bridges.CmdRm(args[2:]...)
	case "show-ref":
		bridges.CmdShowRef(args[0])
	case "status":
		bridges.CmdStatus()
	case "tag":
		var tagObjectFlag bool

		tagObjectCmd := flag.NewFlagSet("tag", flag.ExitOnError)
		tagObjectCmd.BoolVar(&tagObjectFlag, "a", false, "recursively print the tree")

		tagObjectCmd.Parse(args[2:])

		positionalArgs := tagObjectCmd.Args()

		if len(positionalArgs) == 0 {
			bridges.CmdTag("", "", tagObjectFlag)
		} else {
			bridges.CmdTag(positionalArgs[0], positionalArgs[1], tagObjectFlag)
		}
	default:
		log.Fatal("unkown command")
	}
}
