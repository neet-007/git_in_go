package repository

import (
	"fmt"

	"github.com/neet-007/git_in_go/internal/sharedtypes"
	"github.com/neet-007/git_in_go/internal/utils"
)

type GitObject interface {
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
	Init(data []byte)
	GetFmt() ([]byte, error)
}

type GitBlob struct {
	Fmt      string
	BlobData []byte
}

type GitCommit struct {
	Fmt  string
	Kvlm *sharedtypes.Kvlm
}

type GitTree struct {
	Fmt   string
	Items []*GitTreeLeaf
}

type GitTag struct {
	Fmt  string
	Kvlm *sharedtypes.Kvlm
}

func (blob *GitBlob) Serialize() ([]byte, error) {
	return blob.BlobData, nil
}

func (blob *GitBlob) Deserialize(data []byte) error {
	blob.BlobData = data

	return nil
}

func (blob *GitBlob) Init(data []byte) {
	blob.Fmt = "blob"
	blob.BlobData = data
}

func (blob *GitBlob) GetFmt() ([]byte, error) {
	return []byte(blob.Fmt), nil
}

func (commit *GitCommit) Serialize() ([]byte, error) {
	kvlm, err := utils.KvlmSerialize(commit.Kvlm)
	if err != nil {
		return []byte{}, err
	}

	return kvlm, nil
}

func (commit *GitCommit) Deserialize(data []byte) error {
	kvlm, err := utils.KvlmParserWrapper(&data, 0, nil)
	if err != nil {
		return err
	}
	kvlm.Sort()

	commit.Kvlm = kvlm
	return nil
}

func (commit *GitCommit) Init(data []byte) {
	commit.Fmt = "commit"
	err := commit.Deserialize(data)
	if err != nil {
		fmt.Printf("FIX THIS NOT THE WAY TO DO IT BUT GOT ERROR WITH INIT COMMIT:%w", err)
		return
	}
}

func (commit *GitCommit) GetFmt() ([]byte, error) {
	return []byte(commit.Fmt), nil
}

func (tree *GitTree) Serialize() ([]byte, error) {
	return TreeSerialize(tree)
}

func (tree *GitTree) Deserialize(data []byte) error {
	items, err := TreeParser(data)
	if err != nil {
		return err
	}

	tree.Items = items
	return nil
}

func (tree *GitTree) Init(data []byte) {
	tree.Fmt = "tree"
	err := tree.Deserialize(data)
	if err != nil {
		fmt.Printf("FIX THIS NOT THE WAY TO DO IT BUT GOT ERROR WITH INIT TREE:%w\n", err)
		return
	}
}

func (tree *GitTree) GetFmt() ([]byte, error) {
	return []byte(tree.Fmt), nil
}

func (tag *GitTag) Serialize() ([]byte, error) {
	kvlm, err := utils.KvlmSerialize(tag.Kvlm)
	if err != nil {
		return []byte{}, err
	}

	return kvlm, nil
}

func (tag *GitTag) Deserialize(data []byte) error {
	kvlm, err := utils.KvlmParserWrapper(&data, 0, nil)
	if err != nil {
		return err
	}
	kvlm.Sort()

	tag.Kvlm = kvlm
	return nil
}

func (tag *GitTag) Init(data []byte) {
	tag.Fmt = "commit"
	err := tag.Deserialize(data)
	if err != nil {
		fmt.Printf("FIX THIS NOT THE WAY TO DO IT BUT GOT ERROR WITH INIT COMMIT:%w", err)
		return
	}
}

func (tag *GitTag) GetFmt() ([]byte, error) {
	return []byte(tag.Fmt), nil
}
