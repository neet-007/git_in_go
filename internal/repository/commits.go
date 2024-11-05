package repository

import (
	"fmt"
	"strings"
)

func LogGraphviz(repo *Repository, sha string, seen *map[string]byte) error {
	if seen == nil {
		return fmt.Errorf("seen map is nil for commit:%s", sha)
	}

	if _, ok := (*seen)[sha]; ok {
		return nil
	}

	(*seen)[sha] = ' '
	commit, err := repo.ObjectRead(sha)
	if err != nil {
		return fmt.Errorf("could not find repo in loggin for sha:%s", sha)
	}

	commitFmt, err := commit.GetFmt()
	if err != nil {
		return fmt.Errorf("error in fmt for commit with sha:%s", sha)
	}

	if string(commitFmt) != "commit" {
		return fmt.Errorf("the git object does not have type commit for sha:%s type:%s", sha, string(commitFmt))
	}

	gitCommit, ok := commit.(*GitCommit)
	if !ok {
		return fmt.Errorf("expected *GitCommit type but got something else for sha:%s", sha)
	}

	shortHash := sha[:8]

	var messageBytes []byte
	for _, line := range (*(*gitCommit.Kvlm).Map)[""] {
		messageBytes = append(messageBytes, line...)
	}
	message := string(messageBytes)

	message = strings.ReplaceAll(message, "\\", "\\\\")
	message = strings.ReplaceAll(message, "\"", "\\\"")

	if newlineIndex := strings.Index(message, "\n"); newlineIndex != -1 {
		message = message[:newlineIndex]
	}

	fmt.Printf("  c_%s [label=\"%s: %s\"]\n", sha, shortHash, message)

	parents, ok := (*(*gitCommit.Kvlm).Map)["parent"]
	if !ok {
		return nil
	}

	for _, p := range parents {
		parentHash := string(p)
		fmt.Printf("  c_%s -> c_%s;\n", sha, parentHash)

		if err := LogGraphviz(repo, parentHash, seen); err != nil {
			return fmt.Errorf("error in logGraphviz for parent %s: %w", parentHash, err)
		}
	}

	return nil
}
