package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"
)

// MD5Byte -
func MD5Byte(data []byte) string {
	md5 := md5.New()
	md5.Write(data)
	md5Data := md5.Sum([]byte(""))

	return hex.EncodeToString(md5Data)
}

// MD5File -
func MD5File(file *os.File) string {
	md5 := md5.New()
	io.Copy(md5, file)
	md5Data := md5.Sum(nil)

	return hex.EncodeToString(md5Data)
}

// Sha1Byte -
func Sha1Byte(data []byte) string {
	sha1 := sha1.New()
	sha1.Write(data)
	sha1Data := sha1.Sum([]byte(""))

	return hex.EncodeToString(sha1Data)
}

// Sha1File -
func Sha1File(file *os.File) string {
	sha1 := sha1.New()
	io.Copy(sha1, file)
	sha1Data := sha1.Sum(nil)

	return hex.EncodeToString(sha1Data)
}

// Sha1MFile -
func Sha1MFile(file multipart.File) string {
	sha1 := sha1.New()
	io.Copy(sha1, file)
	sha1Data := sha1.Sum(nil)

	return hex.EncodeToString(sha1Data)
}

// PathExists -
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// GetFileSize -
func GetFileSize(path string) (int64, error) {
	//	t := time.Now()
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	//	log.Println(time.Now().Sub(t))
	return size, err
}

// ModifyFileName -
func ModifyFileName(path string, newFileName string) (newPatn string, err error) {
	var tmpPath string
	err = filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			tmpPath = filepath.Join(filepath.Dir(path) + "/" + newFileName)
			os.Rename(path, tmpPath)
		}
		return err
	})

	return tmpPath, err
}

// getFileSize1 -
func getFileSize(filename string) int64 {
	//	t := time.Now()
	var result int64
	filepath.Walk(filename, func(path string, f os.FileInfo, err error) error {
		result = f.Size()
		return nil
	})

	//	log.Println(time.Now().Sub(t))
	return result
}

// getFileSize2 -
func getFileSize2(path string) int64 {
	//	t := time.Now()

	if !PathExists(path) {
		return 0
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0
	}

	//	log.Println(time.Now().Sub(t))
	return fileInfo.Size()
}

// CheckUsernameValidity -
func CheckUsernameValidity(username string) (matched bool) {
	pattern := `^[a-z0-9_-]{3,16}$`
	reg := regexp.MustCompile(pattern)

	return reg.MatchString(username)
}

// CheckEmailValidity -
func CheckEmailValidity(email string) (matched bool) {
	pattern := `^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`
	reg := regexp.MustCompile(pattern)

	return reg.MatchString(email)
}

// CheckPhoneValidity -
func CheckPhoneValidity(phone string) (matched bool) {
	pattern := `^1[3456789]\d{9}$`
	reg := regexp.MustCompile(pattern)

	return reg.MatchString(phone)
}

// GenerateToken - token = md5(username+timestamp+tokenSalt)+timestamp[:]
func GenerateToken(username string) (token string) {
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := MD5Byte([]byte(username + ts + "miuer"))

	return tokenPrefix + ts[:]
}

// ExecLinuxShell -
func ExecLinuxShell(shell string) (err error) {
	cmd := exec.Command("/bin/bash", "-c", shell)

	err = cmd.Run()

	return err
}
