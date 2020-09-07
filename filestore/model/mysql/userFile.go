package mysql

import (
	"database/sql"
	"log"
)

const (
	mysqlCreateUserFileTabel = iota
	mysqlAddNewUserFileMeta
	mysqlGetUserFileMetaByLimit
)

var userFileSQLStrings = []string{
	`CREATE TABLE IF NOT EXISTS user_file(
		id INT(11) NOT NULL AUTO_INCREMENT,
		user_id INT(11) NOT NULL,
		file_sha1 VARCHAR(40) NOT NULL DEFAULT '',
		file_name VARCHAR(256) NOT NULL DEFAULT '',
		file_size BIGINT(20) DEFAULT 0,
		file_addr VARCHAR(256) NOT NULL DEFAULT '',
		create_at TIMESTAMP DEFAULT NOW(),
		update_at TIMESTAMP DEFAULT NOW() ON UPDATE CURRENT_TIMESTAMP(),
		status INT(11) NOT NULL DEFAULT 1,
		PRIMARY KEY(id),
		UNIQUE KEY(file_sha1,file_name),
		KEY(status),
		FOREIGN KEY(file_sha1) REFERENCES public_file(file_sha1) ON DELETE CASCADE,
		FOREIGN KEY(file_addr) REFERENCES public_file(file_addr) ON UPDATE CASCADE,
		FOREIGN KEY(user_id) REFERENCES user(id) ON DELETE CASCADE
	) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4;`,
	`INSERT INTO user_file(user_id,file_sha1, file_name, file_size, file_addr,status)VALUES(?,?,?,?,?,?);`,
	`SELECT file_sha1, file_name, file_size, create_at FROM user_file WHERE userID = ? AND status != 2 ORDER BY create_at DESC LIMIT ?;`,
}

func createUserFileTabel(db *sql.DB) (err error) {
	_, err = db.Exec(userFileSQLStrings[mysqlCreateUserFileTabel])
	return err
}

// AddNewUserFileMeta -
func AddNewUserFileMeta(db *sql.DB, userID int64, fMeta *FileMeta) (err error) {
	stmt, err := db.Prepare(userFileSQLStrings[mysqlAddNewUserFileMeta])
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	res, err := stmt.Exec(userID, fMeta.FileSha1, fMeta.FileName, fMeta.FileSize, fMeta.FileAddr, fMeta.Status)
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

// GetUserFileMetaByLimit -
func GetUserFileMetaByLimit(db *sql.DB, userID int64, limit int64) (fMetas []*FileMeta, err error) {
	stmt, err := db.Prepare(userFileSQLStrings[mysqlGetUserFileMetaByLimit])
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(userID, limit)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	for rows.Next() {
		fMeta := &FileMeta{}

		rows.Scan(&fMeta.FileSha1, &fMeta.FileName, &fMeta.FileSize, &fMeta.CreateAt)

		fMetas = append(fMetas, fMeta)
	}

	return fMetas, nil
}
