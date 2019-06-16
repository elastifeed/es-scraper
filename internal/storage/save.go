package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// FileSaveParams is a simple struct that contains information about how and where to save a file.
type FileSaveParams struct {
	fileType string    // Type (ending) of the file to save
	path     string    // Base path to the folder that should hold all saved files
	folder   uuid.UUID // The deepest level folder, described by a version 5 uuid
}

// New creates a new save struct
func New(fType string) FileSaveParams {
	return FileSaveParams{fileType: fType, path: "/tmp", folder: uuid.New()}
}

// InFolderOf sets save path to the unique folder of the given url.
func (p *FileSaveParams) InFolderOf(url string) *FileSaveParams {
	// Use Version 5 UUIDs that describe the folder based on the request url.
	// This way, request to the same url will be saved in the same folder.
	p.folder = uuid.NewSHA1(uuid.NameSpaceURL, []byte(url))
	return p
}

// Save the given data to the disk and returns the path at which it was saved.
func (p *FileSaveParams) Save(data *[]byte) (string, error) {
	name := p.makeFilename()
	return name, ioutil.WriteFile(name, *data, 0644)

}

// Builds an absolute path to which a file can be saved
func (p *FileSaveParams) makeFilename() string {
	dir := filepath.Join(p.path, p.folder.String())
	_ = os.MkdirAll(dir, os.ModeDir|0700) // Ensure the dir exsits
	// <basePath>/<unique folder based on url>/<filetype>.<filetype>
	return fmt.Sprintf("%s/%s.%s", dir, p.fileType, p.fileType)
}
