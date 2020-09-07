package mysql

import (
	"database/sql"
	"log"
)

const (
	mysqlCreateFileTabel = iota
	mysqlAddNewFileMetaInfo
	mysqlGetPublicFileMeta
	mysqlGetFileMetaBySha1
	mysqlGetFileMetaByLimit
	mysqlUpdateFileMetaBySha1
	mysqlRemoveFileMetaBySha1
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

	// 0-private 1-public 2-forbidden
	// 0-ceph 1-oss
	Status int64
	Ext1   int64
	Ext2   string
}

var fileSQLStrings = []string{
	`CREATE TABLE IF NOT EXISTS public_file(
		id INT(11) NOT NULL AUTO_INCREMENT,
		file_sha1 VARCHAR(40) NOT NULL DEFAULT '',
		file_name VARCHAR(256) NOT NULL DEFAULT '',
		file_size BIGINT(20) DEFAULT 0,
		file_addr VARCHAR(256) NOT NULL DEFAULT '',
		create_at TIMESTAMP DEFAULT NOW(),
		update_at TIMESTAMP DEFAULT NOW() ON UPDATE CURRENT_TIMESTAMP(),
		status INT(11) NOT NULL DEFAULT 1,
		ext1 INT(11) DEFAULT 0,
		ext2 TEXT,
		PRIMARY KEY(id),
		UNIQUE KEY(file_sha1),
		UNIQUE KEY(file_addr),
		KEY(status)
	) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;`,
	`INSERT IGNORE INTO public_file(file_sha1, file_name, file_size, file_addr, status)VALUES(?,?,?,?,?);`,
	`SELECT * FROM public_file WHERE file_sha1 = ? AND status = 1 LIMIT 1;`,
	`SELECT * FROM public_file WHERE file_sha1 = ? LIMIT 1;`,
	`SELECT * FROM public_file WHERE status != 2 ORDER BY create_at DESC LIMIT ?;`,
	`UPDATE public_file SET file_name = ?, file_addr = ? WHERE file_sha1 = ? AND status != 2 LIMIT 1;`,
	`DELETE FROM public_file WHERE file_sha1 = ? LIMIT 1`,
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

	res, err := stmt.Exec(fMeta.FileSha1, fMeta.FileName, fMeta.FileSize, fMeta.FileAddr, fMeta.Status)
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

// GetPublicFileMeta -
func GetPublicFileMeta(db *sql.DB, fSha1 string) (*FileMeta, error) {
	stmt, err := db.Prepare(fileSQLStrings[mysqlGetPublicFileMeta])
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer stmt.Close()

	fMeta := &FileMeta{}

	// including some convert errors, NULL to string
	stmt.QueryRow(fSha1).Scan(
		&fMeta.ID, &fMeta.FileSha1, &fMeta.FileName, &fMeta.FileSize, &fMeta.FileAddr,
		&fMeta.CreateAt, &fMeta.UpdateAt, &fMeta.Status, &fMeta.Ext1, &fMeta.Ext2)

	return fMeta, nil
}

// GetFileMetaBySha1 -
func GetFileMetaBySha1(db *sql.DB, fSha1 string) (*FileMeta, error) {
	stmt, err := db.Prepare(fileSQLStrings[mysqlGetFileMetaBySha1])
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer stmt.Close()

	fMeta := &FileMeta{}

	// including some convert errors, NULL to string
	stmt.QueryRow(fSha1).Scan(
		&fMeta.ID, &fMeta.FileSha1, &fMeta.FileName, &fMeta.FileSize, &fMeta.FileAddr,
		&fMeta.CreateAt, &fMeta.UpdateAt, &fMeta.Status, &fMeta.Ext1, &fMeta.Ext2)

	return fMeta, nil
}

// GetFileMetaByLimit -
func GetFileMetaByLimit(db *sql.DB, limit int64) (fMetas []*FileMeta, err error) {
	stmt, err := db.Prepare(fileSQLStrings[mysqlGetFileMetaByLimit])
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(limit)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	for rows.Next() {
		fMeta := &FileMeta{}

		rows.Scan(&fMeta.ID, &fMeta.FileSha1, &fMeta.FileName, &fMeta.FileSize, &fMeta.FileAddr,
			&fMeta.CreateAt, &fMeta.UpdateAt, &fMeta.Status, &fMeta.Ext1, &fMeta.Ext2)

		fMetas = append(fMetas, fMeta)
	}

	return fMetas, nil
}

// UpdateFileMetaBySha1 -
func UpdateFileMetaBySha1(db *sql.DB, fileName, fileAddr, fileSha1 string) (err error) {
	stmt, err := db.Prepare(fileSQLStrings[mysqlUpdateFileMetaBySha1])
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(fileName, fileAddr, fileSha1)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// RemoveFileMetaBySha1 -
func RemoveFileMetaBySha1(db *sql.DB, fileSha1 string) (err error) {
	stmt, err := db.Prepare(fileSQLStrings[mysqlRemoveFileMetaBySha1])
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(fileSha1)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
