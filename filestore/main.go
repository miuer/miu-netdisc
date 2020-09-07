package main

import (
	"fmt"
	"log"
	"os"

	"github.com/miuer/miu-netdisc/filestore/config"

	"github.com/miuer/miu-netdisc/filestore/model/transfer"

	"github.com/miuer/miu-netdisc/filestore/model/ceph"
	"github.com/miuer/miu-netdisc/filestore/model/rabbitmq"

	"github.com/miuer/miu-netdisc/filestore/controller/handler"
	"github.com/miuer/miu-netdisc/filestore/model/mysql"
)

func main() {
	fmt.Println("Hollow World!")

	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Llongfile)

	os.MkdirAll(config.TmpDataFileDir, 0744)
	os.MkdirAll(config.TmpChunkFileDir, 0744)

	writer, reader := mysql.InitMysql()
	//rdsPool := rds.InitCache()
	cephConn := ceph.InitCeph()

	go rabbitmq.Consume(writer, transfer.Transfer)

	handler.InitRouter(writer, reader, cephConn)
}
