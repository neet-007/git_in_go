package repository

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
	Fmt      string
	BlobData []byte
}

type GitTree struct {
	Fmt      string
	BlobData []byte
}

type GitTag struct {
	Fmt      string
	BlobData []byte
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

func (blob *GitCommit) Serialize() ([]byte, error) {
	return blob.BlobData, nil
}
func (blob *GitCommit) Deserialize(data []byte) error {
	blob.BlobData = data

	return nil
}
func (blob *GitCommit) Init(data []byte) {

}
func (blob *GitCommit) GetFmt() ([]byte, error) {
	return []byte(blob.Fmt), nil
}

func (blob *GitTree) Serialize() ([]byte, error) {
	return blob.BlobData, nil
}
func (blob *GitTree) Deserialize(data []byte) error {
	blob.BlobData = data

	return nil
}
func (blob *GitTree) Init(data []byte) {

}
func (blob *GitTree) GetFmt() ([]byte, error) {
	return []byte(blob.Fmt), nil
}

func (blob *GitTag) Serialize() ([]byte, error) {
	return blob.BlobData, nil
}
func (blob *GitTag) Deserialize(data []byte) error {
	blob.BlobData = data

	return nil
}
func (blob *GitTag) Init(data []byte) {

}
func (blob *GitTag) GetFmt() ([]byte, error) {
	return []byte(blob.Fmt), nil
}
