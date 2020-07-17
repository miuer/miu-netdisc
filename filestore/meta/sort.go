package meta

import (
	"time"
)

type byUploadAt []FileMeta

func (fm byUploadAt) Len() int {
	return len(fm)
}

func (fm byUploadAt) Swap(i, j int) {
	fm[i], fm[j] = fm[j], fm[i]
}

func (fm byUploadAt) Less(i, j int) bool {
	iTime, _ := time.Parse("2006-01-02 15:04:05", fm[i].UploadAt)
	jTime, _ := time.Parse("2006-01-02 15:04:05", fm[j].UploadAt)

	return iTime.UnixNano() > jTime.UnixNano()
}
