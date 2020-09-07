package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/miuer/miu-netdisc/filestore/model/rabbitmq"

	"github.com/miuer/miu-netdisc/filestore/config"

	"github.com/miuer/miu-netdisc/filestore/model/ceph"

	"github.com/miuer/miu-netdisc/filestore/model/rds"

	osss "github.com/miuer/miu-netdisc/filestore/model/oss"

	"github.com/miuer/miu-netdisc/filestore/model/mysql"
	"github.com/miuer/miu-netdisc/filestore/utils"
)

func (ctl *Controller) uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("welcome to file store system."))
		return
	} else if r.Method == http.MethodPost {
		userID := r.FormValue("userID")
		iuserID, _ := strconv.ParseInt(userID, 10, 64)
		fileStatus := r.FormValue("fileStatus")
		ifileStatus, _ := strconv.ParseInt(fileStatus, 10, 64)

		file, header, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "Failed to get file data, err:"+err.Error())
			return
		}

		f, err := ioutil.ReadAll(file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Failed to convert file to []byte, err:"+err.Error())
			return
		}

		fMeta := &mysql.FileMeta{
			FileName: header.Filename,
			FileAddr: config.TmpDataFileDir + header.Filename,
			FileSize: header.Size,
			CreateAt: time.Now().Format("2006-01-02 15:04:05"),
			FileSha1: utils.Sha1Byte(f),
			Status:   ifileStatus,
		}

		defer file.Close()

		fm, err := mysql.GetFileMetaBySha1(ctl.Reader, fMeta.FileSha1)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Failed to get file meta, err:"+err.Error())
			return
		}

		if fm.ID > 0 {
			// fast upload

			fm.FileName = header.Filename
			mysql.AddNewUserFileMeta(ctl.Writer, iuserID, fm)
			http.Redirect(w, r, "/file/fastUploadSucceed", http.StatusFound)
			return
		}

		if fMeta.FileSize > (5 * 1024 * 1024) {

			ctl.chunkUpload(w, r)
			return
		}

		newFile, err := os.Create(fMeta.FileAddr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Failed to create file, err:"+err.Error())
			return
		}

		defer newFile.Close()

		_, err = newFile.Write(f)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Failed to save data into file, err:"+err.Error())
			return
		}

		//	newFile.Seek(0, 0)
		//	fMeta.FileSha1 = utils.Sha1File(newFile)

		// ---- puplic - user

		if fMeta.Status == 0 {
			// ceph 同步
			bucket := ceph.GetBucket(ctl.CephConn)
			cephPath := config.CephRootPath + fMeta.FileSha1 + "/" + fMeta.FileName
			err := ceph.TransferToCeph(bucket, cephPath, f)
			if err != nil {

			}

			fMeta.FileAddr = cephPath

		} else {
			destPath := config.OssRootPath + fMeta.FileSha1 + "/" + fMeta.FileName
			msg := rabbitmq.TransferMeta{
				FileName:     fMeta.FileName,
				FileSha1:     fMeta.FileSha1,
				FileSize:     fMeta.FileSize,
				FileCurAddr:  fMeta.FileAddr,
				FileDestAddr: destPath,
			}

			pubData, _ := json.Marshal(msg)

			err = rabbitmq.Publish(pubData)
			if err != nil {
				log.Println("转移oss失败")
			}

		}

		mysql.AddNewFileMeta(ctl.Writer, fMeta)
		mysql.AddNewUserFileMeta(ctl.Writer, iuserID, fMeta)
	}

	http.Redirect(w, r, "/file/uploadSucceed", http.StatusFound)
}

func uploadSucceedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("Uploaded Successfully"))
}

func fastUploadSucceedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("FastUploaded Successfully"))
}

func (ctl *Controller) getFileMetaHandler(w http.ResponseWriter, r *http.Request) {

	fileSha1 := r.FormValue("fileSha1")

	fMeta, err := mysql.GetPublicFileMeta(ctl.Reader, fileSha1)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}

	data, err := json.Marshal(fMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Failed to convert data to json, err:"+err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (ctl *Controller) updateFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	op := r.FormValue("op")
	fileSha1 := r.FormValue("fileSha1")
	newFileName := r.FormValue("fileName")

	if op != "1" {
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, errors.New("No permission to perform update file meta").Error())
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	curFMeta, err := mysql.GetFileMetaBySha1(ctl.Reader, fileSha1)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}

	curFMeta.FileName = newFileName
	curFMeta.FileAddr, err = utils.ModifyFileName(curFMeta.FileAddr, newFileName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Failed to update file path, err:"+err.Error())
		return
	}

	err = mysql.UpdateFileMetaBySha1(ctl.Writer, curFMeta.FileName, curFMeta.FileAddr, curFMeta.FileSha1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, err.Error())
		return
	}

	data, err := json.Marshal(curFMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Failed to convert data to json, err:"+err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (ctl *Controller) queryHandler(w http.ResponseWriter, r *http.Request) {
	limit := r.FormValue("limit")

	ilimit, _ := strconv.ParseInt(limit, 0, 64)

	fileMetas, err := mysql.GetFileMetaByLimit(ctl.Reader, ilimit)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}

	data, err := json.Marshal(fileMetas)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Failed to convert data to json, err:"+err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (ctl *Controller) downloadHandler(w http.ResponseWriter, r *http.Request) {

	fileSha1 := r.FormValue("fileSha1")

	fMeta, err := mysql.GetFileMetaBySha1(ctl.Reader, fileSha1)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}

	var data []byte

	if strings.HasPrefix(fMeta.FileAddr, config.OssRootPath) {
		url := osss.DownloadURL(fMeta.FileAddr)

		io.WriteString(w, url)
		return
	} else if strings.HasPrefix(fMeta.FileAddr, config.CephRootPath) {
		bucket := ceph.GetBucket(ctl.CephConn)
		data, err = bucket.Get(fMeta.FileAddr)
		if err != nil {
			log.Println("get ceph file error")
		}

	} else {

		file, err := os.Open(fMeta.FileAddr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Failed to open file in server, err:"+err.Error())
			return
		}
		defer file.Close()

		data, err = ioutil.ReadAll(file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Failed to download file, err:"+err.Error())
			return
		}

	}
	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("content-disposition", "attachment;filename=\""+fMeta.FileName+"\"")

	w.Write(data)
}

func (ctl *Controller) deleteHandler(w http.ResponseWriter, r *http.Request) {
	fileSha1 := r.FormValue("fileSha1")

	fMeta, err := mysql.GetFileMetaBySha1(ctl.Reader, fileSha1)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}

	err = os.Remove(fMeta.FileAddr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Failed to delete file, err:"+err.Error())
		return
	}

	err = mysql.RemoveFileMetaBySha1(ctl.Writer, fileSha1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Failed to delete file meta, err:"+err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("file successfully deleted"))
}

func (ctl *Controller) chunkUpload(w http.ResponseWriter, r *http.Request) {
	// init

	userID := r.FormValue("userID")
	iuserID, _ := strconv.ParseInt(userID, 10, 64)

	file, header, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Failed to get file data, err:"+err.Error())
		return
	}

	f, err := ioutil.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Failed to get src file, err:"+err.Error())
		return
	}

	chunkInfo := &rds.ChunkInfo{
		FileSha1:   utils.Sha1Byte(f),
		FileSize:   header.Size,
		ChunkSize:  rds.ChunkSize,
		ChunkCount: int64(math.Ceil(float64(header.Size) / float64(rds.ChunkSize))),
	}

	// block upload

	fileBuf := bytes.NewReader(f)
	chunkBuf := make([]byte, chunkInfo.ChunkSize)

	index := 0
	ch := make(chan int)

	chunkFileDir := config.TmpChunkFileDir + chunkInfo.FileSha1 + "/"
	dataFileDir := config.TmpDataFileDir
	chunkInspectionFilePath := config.TmpChunkFileDir + chunkInfo.FileSha1 + ".txt"

	ckInsFile, err := os.OpenFile(chunkInspectionFilePath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Failed to open file, err:"+err.Error())
		return
	}

	mux := &sync.RWMutex{}

	if utils.PathExists(chunkFileDir) {
		// duandian

		insData, err := ioutil.ReadFile(chunkInspectionFilePath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Failed to get chunk inspection file, err:"+err.Error())
			return
		}

		insLen := strings.Split(string(insData), "/")

		if (len(insLen) - 1) != int(chunkInfo.ChunkCount) {

			for {
				n, err := fileBuf.Read(chunkBuf)
				if err != nil {
					if err == io.EOF {
						break
					} else {
						w.WriteHeader(http.StatusInternalServerError)
						io.WriteString(w, "Failed to convert file to []byte, err:"+err.Error())
						return
					}
				}

				if n <= 0 {
					break
				}

				index++

				bufCopied := make([]byte, chunkInfo.ChunkSize)
				copy(bufCopied, chunkBuf)

				if strings.Contains(string(insData), utils.Sha1Byte(bufCopied[:n])) {
					continue
				}

				go func(bt []byte, curIDX int, mux *sync.RWMutex) {
					filePath := chunkFileDir + strconv.Itoa(curIDX)
					chunkFile, err := os.Create(filePath)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						io.WriteString(w, "Failed to create chunk file, err:"+err.Error())
						return
					}

					defer chunkFile.Close()

					_, err = chunkFile.Write(bt)
					if err != nil {
						w.WriteHeader(http.StatusInternalServerError)
						io.WriteString(w, "Failed to write chunk file, err:"+err.Error())
						return
					}

					mux.Lock()
					ckInsFile.Seek(0, io.SeekEnd)
					ckInsFile.WriteString(utils.Sha1Byte(bt) + "/")
					mux.Unlock()

					ch <- curIDX

				}(bufCopied[:n], index, mux)

			}
		}

	} else {

		//	os.MkdirAll(chunkFileDir, 0744)
		//	os.MkdirAll(dataFileDir, 0744)

		for {
			n, err := fileBuf.Read(chunkBuf)
			if err != nil {
				if err == io.EOF {
					break
				} else {
					w.WriteHeader(http.StatusInternalServerError)
					io.WriteString(w, "Failed to convert file to []byte, err:"+err.Error())
					return
				}
			}

			if n <= 0 {
				break
			}

			index++

			bufCopied := make([]byte, chunkInfo.ChunkSize)
			copy(bufCopied, chunkBuf)

			go func(bt []byte, curIDX int, mux *sync.RWMutex) {
				filePath := chunkFileDir + strconv.Itoa(curIDX)
				chunkFile, err := os.Create(filePath)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					io.WriteString(w, "Failed to create chunk file, err:"+err.Error())
					return
				}

				defer chunkFile.Close()

				_, err = chunkFile.Write(bt)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					io.WriteString(w, "Failed to write chunk file, err:"+err.Error())
					return
				}

				mux.Lock()
				ckInsFile.Seek(0, io.SeekEnd)
				ckInsFile.WriteString(utils.Sha1Byte(bt) + "/")
				mux.Unlock()

				ch <- curIDX

			}(bufCopied[:n], index, mux)

		}

	}
	// concurrency control
	for idx := 0; idx < index; idx++ {
		select {
		case <-ch:
		}
	}

	ckInsFile.Close()

	// merge

	insData, err := ioutil.ReadFile(chunkInspectionFilePath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Failed to write chunk file, err:"+err.Error())
		return
	}

	insLen := strings.Split(string(insData), "/")

	if (len(insLen) - 1) != int(chunkInfo.ChunkCount) {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Failed to upload file,err: chunk file is missing")
		return
	}

	srcDir := chunkFileDir
	destPath := dataFileDir + header.Filename

	shell := fmt.Sprintf("cd %s && ls | sort -n | xargs cat > %s", srcDir, destPath)

	err = utils.ExecLinuxShell(shell)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Failed to meger chunk file,err:"+err.Error())
		return
	}

	fMeta := &mysql.FileMeta{
		FileName: header.Filename,
		FileAddr: config.TmpDataFileDir + header.Filename,
		FileSize: header.Size,
		CreateAt: time.Now().Format("2006-01-02 15:04:05"),
		FileSha1: utils.Sha1Byte(f),
	}

	mysql.AddNewFileMeta(ctl.Writer, fMeta)
	mysql.AddNewUserFileMeta(ctl.Writer, iuserID, fMeta)

	w.WriteHeader(http.StatusOK)

	w.Write([]byte("ChunkUloaded Successfully"))
}
