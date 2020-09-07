package mysql_test

import (
	"log"
	"testing"

	"github.com/miuer/miu-netdisc/filestore/model/mysql"
)

func TestDemo(t *testing.T) {
	_, reader := mysql.InitMysql()

	id, err := mysql.GetIDByToken(reader, "834ca987d9350578650a69166876509b5f1915cb")

	log.Println(id, err)
}
