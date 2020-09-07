package ceph

import (
	"github.com/miuer/miu-netdisc/filestore/config"
	"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"
)

// InitCeph -
func InitCeph() *s3.S3 {
	conn := s3.New(
		aws.Auth{
			AccessKey: "I2OMC6XU3DX6BLW31H2N",
			SecretKey: "vNhv0IpCSnKr8WXKf878wEB0lPts4rgnK5RzSsrR",
		},
		aws.Region{
			Name:                 "miuer",
			EC2Endpoint:          "http://" + config.CephAddr,
			S3Endpoint:           "http://" + config.CephAddr,
			S3BucketEndpoint:     "",
			S3LocationConstraint: false,
			S3LowercaseBucket:    false,
			Sign:                 aws.SignV2,
		},
	)

	return conn
}

// GetBucket -
func GetBucket(conn *s3.S3) (bucket *s3.Bucket) {
	bucket = conn.Bucket(config.CephUser)
	bucket.PutBucket(s3.Private)
	return bucket
}

// TransferToCeph -
func TransferToCeph(bucket *s3.Bucket, cephPath string, data []byte) (err error) {
	return bucket.Put(cephPath, data, "application/octet-stream", s3.Private)
}
