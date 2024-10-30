package repository_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"os/exec"
	"testing"

	"github.com/neet-007/git_in_go/internal/repository"
)

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func TestHashObjectBlob(t *testing.T) {
	cases := make([]string, 10, 10)

	for range len(cases) {
		cases = append(cases, randomString(10))
	}

	for _, str := range cases {
		obj := repository.GitBlob{}
		obj.Init([]byte(str))

		myHash, err := repository.ObjectWrite(&obj, nil)
		if err != nil {
			t.Fatalf("Failed to hash with custom function: %v", err)
		}
		fmt.Printf("My custom hash: %s\n", myHash)

		cmd := exec.Command("git", "hash-object", "-t", "blob", "--stdin")
		cmd.Stdin = bytes.NewReader([]byte(str))

		gitHash, err := cmd.Output()
		if err != nil {
			t.Fatalf("Failed to run git hash-object: %v", err)
		}

		gitHashStr := string(bytes.TrimSpace(gitHash))
		fmt.Printf("Git hash-object hash: %s\n", gitHashStr)

		if myHash == gitHashStr {
			fmt.Println("Hashes match!")
		} else {
			t.Fatalf("Hashes do not match! exp %s vs got %s\n", gitHashStr, myHash)
		}
	}
}
