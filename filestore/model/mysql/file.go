package mysql

import (
	"database/sql"
	"log"
)

const (
	mysqlCreateFileTabel = iota
	mysqlAddNewFileMetaInfo
)

// FileMeta -
type FileMeta struct {
	ID       int64
	FileSha1 string
	FileName string
	FileSize int64
	FileAddr string
	CreateAt string
	UpdateAt string
	Status   int64
	Ext1     int64
	Ext2     string
}

var fileSQLStrings = []string{
	`CREATE TABLE IF NOT EXISTS public_file(
		id INT(11) NOT NULL AUTO_INCREMENT,
		file_sha1 VARCHAR(40) NOT NULL DEFAULT '',
		file_name VARCHAR(256) NOT NULL DEFAULT '',
		file_size BIGINT(20) DEFAULT 0,
		file_addr VARCHAR(1024) NOT NULL DEFAULT '',
		create_at TIMESTAMP DEFAULT NOW(),
		update_at TIMESTAMP DEFAULT NOW() ON UPDATE CURRENT_TIMESTAMP(),
		status INT(11) NOT NULL DEFAULT 1,
		ext1 INT(11) DEFAULT 0,
		ext2 TEXT,
		PRIMARY KEY(id),
		UNIQUE KEY(file_sha1),
		KEY(status)
	) ENGINE = InnoDB DEFAULT CHARSET = utf8`,
	`INSERT IGNORE INTO public_file(file_sha1, file_name, file_size, file_addr)VALUES(?,?,?,?);`,
}

// CreateFileTable -
func createFileTable(db *sql.DB) (err error) {
	_, err = db.Exec(fileSQLStrings[mysqlCreateFileTabel])
	return err
}

// AddNewFileMeta -
func AddNewFileMeta(db *sql.DB, fMeta *FileMeta) (err error) {
	stmt, err := db.Prepare(fileSQLStrings[mysqlAddNewFileMetaInfo])
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(fMeta.FileSha1, fMeta.FileName, fMeta.FileSize, fMeta.FileAddr)
	if err != nil {
		log.Println(err)
		return err
	}

	rf, err := res.RowsAffected()
	if err == nil {
		if rf <= 0 {
			log.Printf("file %s has been uploaded before\n", fMeta.FileSha1)
		}
	}

	return err
}
