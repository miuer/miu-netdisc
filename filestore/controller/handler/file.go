package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/miuer/miu-netdisc/filestore/meta"
	"github.com/miuer/miu-netdisc/filestore/model/mysql"
	"github.com/miuer/miu-netdisc/filestore/utils"
)

// Controller -
type Controller struct {
	Writer *sql.DB
	Reader *sql.DB
}

func (ctl *Controller) uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("welcome to file store system."))
		return
	} else if r.Method == "POST" {
		r.ParseForm()

		file, head, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "Failed to get file data, err:"+err.Error())
			return
		}

		fMeta := &mysql.FileMeta{
			FileName: head.Filename,
			FileAddr: "/Users/duanyahong/tmp/" + head.Filename,
			CreateAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		defer file.Close()

		newFile, err := os.Create(fMeta.FileAddr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Failed to create file, err:"+err.Error())
			return
		}

		defer newFile.Close()

		fMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, "Failed to save data into file, err:"+err.Error())
			return
		}

		newFile.Seek(0, 0)
		fMeta.FileSha1 = utils.Sha1File(newFile)

		mysql.AddNewFileMeta(ctl.Writer, fMeta)
	}

	http.Redirect(w, r, "/file/uploadSucceed", http.StatusFound)
}

func uploadSucceedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("Uploaded Successfully"))
}

func getFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	fileSha1 := r.Form["fileSha1"][0]

	fMeta, err := meta.GetFileMeta(fileSha1)
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

func updateFileMetaHandler(w http.ResponseWriter, r *http.Request) {
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

	curFMeta, err := meta.GetFileMeta(fileSha1)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}

	curFMeta.FileName = newFileName
	// curFMeta.Location = "/Users/duanyahong/tmp/" + newFileName

	err = curFMeta.UpdateFileMeta()
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

func queryHandler(w http.ResponseWriter, r *http.Request) {
	limit := r.FormValue("limit")

	ilimit, _ := strconv.Atoi(limit)

	fileMetas, err := meta.GetLastFileMetas(ilimit)
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

func downloadHandler(w http.ResponseWriter, r *http.Request) {

	fileSha1 := r.FormValue("fileSha1")

	fMeta, err := meta.GetFileMeta(fileSha1)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}

	file, err := os.Open(fMeta.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Failed to open file in server, err:"+err.Error())
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Failed to download file, err:"+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("content-disposition", "attachment;filename=\""+fMeta.FileName+"\"")

	w.Write(data)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	fileSha1 := r.FormValue("fileSha1")

	fMeta, err := meta.GetFileMeta(fileSha1)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, err.Error())
		return
	}

	err = os.Remove(fMeta.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Failed to delete file, err:"+err.Error())
		return
	}

	err = meta.RemoveFileMeta(fileSha1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		io.WriteString(w, "Failed to delete file meta, err:"+err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("file successfully deleted"))
}
