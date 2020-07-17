package handler

import (
	"database/sql"
	"net/http"
)

// InitRouter -
func InitRouter(writer *sql.DB, reader *sql.DB) {

	fileCtl := &Controller{
		Writer: writer,
		Reader: reader,
	}

	http.HandleFunc("/file/upload", fileCtl.uploadHandler)
	http.HandleFunc("/file/uploadSucceed", uploadSucceedHandler)
	http.HandleFunc("/meta/getFileMeta", getFileMetaHandler)
	http.HandleFunc("/meta/updateFileMeta", updateFileMetaHandler)
	http.HandleFunc("/file/query", queryHandler)
	http.HandleFunc("/file/download", downloadHandler)
	http.HandleFunc("/file/delete", deleteHandler)

	http.ListenAndServe("127.0.0.1:18080", nil)
}
