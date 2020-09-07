package mysql

import (
	"database/sql"
	"log"
)

const (
	mysqlCreateUserTabel = iota
	mysqlAddNewUser
	mysqlGetIDAndPwdByUsername
	mysqlGetUsernameByID
	mysqlGetIDByEmail
	mysqlGetIDByPhone
)

// User -
type User struct {
	ID       int64
	Username string
	Password string
	Email    string
	Phone    string
	CreateAt string
	Status   int64
}

var userSQLStrings = []string{
	`CREATE TABLE IF NOT EXISTS user(
		id INT(11) NOT NULL AUTO_INCREMENT,
		username VARCHAR(64) NOT NULL DEFAULT "",
		password VARCHAR(128) NOT NULL DEFAULT "",
		email VARCHAR(256) NOT NULL DEFAULT "",
		phone VARCHAR(256) NOT NULL DEFAULT "",
		create_at TIMESTAMP DEFAULT NOW(),
		status INT(11) NOT NULL DEFAULT 0,
		PRIMARY KEY(id),
		UNIQUE KEY(username),
		UNIQUE KEY(email),
		UNIQUE KEY(phone),
		KEY(status)
	) ENGINE = InnoDB AUTO_INCREMENT = 1000 DEFAULT CHARSET = utf8mb4;`,
	`INSERT INTO user(username, password, email, phone)VALUES(?, ?, ?, ?);`,
	`SELECT id, password FROM user WHERE username = ? limit 1;`,
	`SELECT username FROM user WHERE id = ? limit 1;`,
	`SELECT id FROM user WHERE email = ? limit 1;`,
	`SELECT id FROM user WHERE phone = ? limit 1;`,
}

// createUserTable -
func createUserTable(db *sql.DB) (err error) {
	_, err = db.Exec(userSQLStrings[mysqlCreateUserTabel])
	return err
}

// AddNewUser -
func AddNewUser(db *sql.DB, user *User) (err error) {
	stmt, err := db.Prepare(userSQLStrings[mysqlAddNewUser])
	if err != nil {
		log.Println(err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.Username, user.Password, user.Email, user.Phone)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// GetIDAndPwdByUsername -
func GetIDAndPwdByUsername(db *sql.DB, username string) (id int64, pwd string, err error) {
	stmt, err := db.Prepare(userSQLStrings[mysqlGetIDAndPwdByUsername])
	if err != nil {
		log.Println(err)
		return 0, "", err
	}
	defer stmt.Close()

	stmt.QueryRow(username).Scan(&id, &pwd)

	return id, pwd, nil
}

// GetUsernameByID -
func GetUsernameByID(db *sql.DB, id int64) (username string, err error) {
	stmt, err := db.Prepare(userSQLStrings[mysqlGetUsernameByID])
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer stmt.Close()

	stmt.QueryRow(id).Scan(&username)

	return username, nil
}

// GetIDByEmail -
func GetIDByEmail(db *sql.DB, email string) (id int64, err error) {
	stmt, err := db.Prepare(userSQLStrings[mysqlGetIDByEmail])
	if err != nil {
		log.Println(err)
		return 0, err
	}
	defer stmt.Close()

	stmt.QueryRow(email).Scan(&id)

	return id, nil
}

// GetIDByPhone -
func GetIDByPhone(db *sql.DB, phone string) (id int64, err error) {
	stmt, err := db.Prepare(userSQLStrings[mysqlGetIDByPhone])
	if err != nil {
		log.Println(err)
		return 0, err
	}
	defer stmt.Close()

	stmt.QueryRow(phone).Scan(&id)

	return id, nil
}
