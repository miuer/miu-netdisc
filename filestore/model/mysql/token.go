package mysql

import (
	"database/sql"
	"log"
)

const (
	mysqlCreateTokenTabel = iota
	mysqlReplaceToken
	mysqlGetIDByToken
)

var tokenSQLStrings = []string{
	`CREATE TABLE IF NOT EXISTS user_token(
		id INT(11) NOT NULL AUTO_INCREMENT,
		user_id INT(11) NOT NULL,
		token CHAR(40) NOT NULL DEFAULT "",
		PRIMARY KEY(id),
		UNIQUE KEY(user_id),
		KEY(token),
		FOREIGN KEY(user_id) REFERENCES user(id) ON DELETE CASCADE
	) ENGINE = InnoDB DEFAULT CHARSET = utf8;`,
	`REPLACE INTO user_token(user_id, token)VALUES(?,?);`,
	`SELECT id FROM user_token WHERE token = ? limit 1;`,
}

// createTokenTable -
func createTokenTable(db *sql.DB) (err error) {
	_, err = db.Exec(tokenSQLStrings[mysqlCreateTokenTabel])
	return err
}

// ReplaceToken -
func ReplaceToken(db *sql.DB, id int64, token string) (err error) {
	stmt, err := db.Prepare(tokenSQLStrings[mysqlReplaceToken])
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id, token)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// GetIDByToken -
func GetIDByToken(db *sql.DB, token string) (id int64, err error) {
	stmt, err := db.Prepare(tokenSQLStrings[mysqlGetIDByToken])
	if err != nil {
		log.Println(err)
		return 0, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(token).Scan(&id)

	return id, err
}
