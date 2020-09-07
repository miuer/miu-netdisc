package oss

import (
	"bufio"
	"fmt"
	"os"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/miuer/miu-netdisc/filestore/config"
	"github.com/miuer/miu-netdisc/filestore/model/rabbitmq"
	"github.com/miuer/miu-netdisc/filestore/model/rds"
)

// InitOss -
func initOss() (client *oss.Client) {

	client, _ = oss.New(config.OssEndPoint, config.OssAccessKeyID, config.OssAccessKeySecret)

	return client
}

// GetBucket -
func getBucket() (bucket *oss.Bucket, err error) {
	client := initOss()
	err = client.CreateBucket(config.OssBucketName, oss.ACL(oss.ACLPrivate))
	if err != nil {
		return nil, err
	}

	bucket, err = client.Bucket(config.OssBucketName)
	return bucket, err
}

// TransferToOss -
func TransferToOss(tMeta *rabbitmq.TransferMeta) (err error) {
	bucket, err := getBucket()
	if err != nil {
		return err
	}

	fd, err := os.Open(tMeta.FileCurAddr)
	if err != nil {
		return err
	}
	defer fd.Close()

	if tMeta.FileSize < rds.ChunkSize {
		// 简单上传
		err = bucket.PutObject(
			tMeta.FileDestAddr,
			bufio.NewReader(fd),
			oss.ACL(oss.ACLPrivate),
		)
		return err
	}

	chunks, err := oss.SplitFileByPartNum(tMeta.FileCurAddr, 3)
	if err != nil {
		return err
	}

	storageType := oss.ObjectStorageClass(oss.StorageStandard)
	imur, err := bucket.InitiateMultipartUpload(tMeta.FileDestAddr, storageType)
	if err != nil {
		return err
	}

	var parts []oss.UploadPart
	for _, chunk := range chunks {
		fd.Seek(chunk.Offset, os.SEEK_SET)
		part, err := bucket.UploadPart(imur, fd, chunk.Size, chunk.Number)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(-1)
		}
		parts = append(parts, part)
	}

	_, err = bucket.CompleteMultipartUpload(imur, parts)

	return err

}

// DownloadURL -
func DownloadURL(fileAddr string) (url string) {
	bucket, err := getBucket()
	if err != nil {
		return ""
	}

	url, err = bucket.SignURL(fileAddr, oss.HTTPGet, 3600)
	if err != nil {
		return ""
	}

	return url
}
