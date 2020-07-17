package meta

import (
	"errors"
	"sort"
	"sync"
)

// FileMeta -
type FileMeta struct {
	FileSha1 string
	FileName string
	FileSize int64
	Location string
	UploadAt string
}

var fileMetas map[string]*FileMeta
var mu *sync.Mutex

func init() {
	fileMetas = make(map[string]*FileMeta)
	mu = new(sync.Mutex)
}

// UpdateFileMeta -
func (fmeta *FileMeta) UpdateFileMeta() (err error) {
	fileMetas[fmeta.FileSha1] = fmeta
	return nil
}

// GetFileMeta -
func GetFileMeta(fileSha1 string) (fmeta *FileMeta, err error) {
	fmeta = fileMetas[fileSha1]

	if fmeta == nil {
		return nil, errors.New("Failed to get file meta, err: file meta not exists")
	}

	return fmeta, nil
}

// GetLastFileMetas -
func GetLastFileMetas(count int) (fMetas []FileMeta, err error) {

	fMetaArray := make([]FileMeta, len(fileMetas))

	if cap(fMetaArray) < count {
		return nil, errors.New("Failed to get file meta, err: out of range count")
	}

	for _, v := range fileMetas {
		fMetaArray = append(fMetaArray, *v)
	}

	sort.Sort(byUploadAt(fMetaArray))

	return fMetaArray[:count], nil
}

// RemoveFileMeta -
func RemoveFileMeta(fileSha1 string) (err error) {
	mu.Lock()
	delete(fileMetas, fileSha1)
	mu.Unlock()

	return nil
}
