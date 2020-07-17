package main

import (
	"fmt"

	"github.com/miuer/miu-netdisc/filestore/model/mysql"

	"github.com/miuer/miu-netdisc/filestore/handler"
)

func main() {
	fmt.Println("Hollow World!")

	handler.InitRouter(	mysql.InitMysql()
)
}
