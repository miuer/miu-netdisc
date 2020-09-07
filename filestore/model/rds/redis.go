package rds

const (
	// ChunkSize -
	ChunkSize = 5 * 1024 * 1024
)

// ChunkInfo -
type ChunkInfo struct {
	FileSha1   string
	FileSize   int64
	ChunkSize  int64
	ChunkCount int64
}
