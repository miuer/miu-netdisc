package main

import (
	"github.com/miuer/miu-netdisc/filestore/model/ceph"
	"github.com/miuer/miu-netdisc/filestore/model/mysql"
	controller "github.com/miuer/miu-netdisc/micro-storage/controller/gin"
)

func main() {
	writer, reader := mysql.InitMysql()
	//rdsPool := rds.InitCache()
	cephConn := ceph.InitCeph()

	ctl := &controller.Controller{
		writer,
		reader,
		cephConn,
	}

	r := controller.InitRouter(ctl)

	r.Run()
}
