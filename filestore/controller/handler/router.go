package handler

import (
	"database/sql"
	"net/http"

	"gopkg.in/amz.v1/s3"
)

// Controller -
type Controller struct {
	Writer *sql.DB
	Reader *sql.DB
	//	RdsPool *redis.Pool
	CephConn *s3.S3
}

// InitRouter -
func InitRouter(writer *sql.DB, reader *sql.DB, cephConn *s3.S3) {

	fileCtl := &Controller{
		Writer: writer,
		Reader: reader,
		//	RdsPool: rdsPool,
		CephConn: cephConn,
	}

	userCtl := &Controller{
		Writer: writer,
		Reader: reader,
	}

	// 	http.HandleFunc("/file/upload", fileCtl.uploadHandler)
	http.HandleFunc("/file/uploadSucceed", uploadSucceedHandler)
	http.HandleFunc("/file/fastUploadSucceed", fastUploadSucceedHandler)
	http.HandleFunc("/meta/getFileMeta", fileCtl.getFileMetaHandler)
	http.HandleFunc("/meta/updateFileMeta", fileCtl.updateFileMetaHandler)
	http.HandleFunc("/file/query", fileCtl.queryHandler)
	http.HandleFunc("/file/download", fileCtl.downloadHandler)
	http.HandleFunc("/file/delete", fileCtl.deleteHandler)

	http.HandleFunc("/file/testChunk", fileCtl.chunkUpload)

	http.HandleFunc("/user/register", userCtl.RegisterHandler)
	http.HandleFunc("/user/registerSucceed", userCtl.RegisterSucceedHandler)
	http.HandleFunc("/user/login", userCtl.LoginHandler)
	http.Handle("/file/upload", userCtl.CheckTokenValidity(http.HandlerFunc(fileCtl.uploadHandler)))
	//	http.Handle("/file/download", userCtl.CheckTokenValidity(http.HandlerFunc(fileCtl.downloadHandler)))

	http.ListenAndServe(":18080", nil)
}
