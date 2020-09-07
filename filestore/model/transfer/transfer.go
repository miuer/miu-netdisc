package transfer

import (
	"database/sql"
	"encoding/json"

	"github.com/miuer/miu-netdisc/filestore/model/mysql"

	"github.com/miuer/miu-netdisc/filestore/model/rabbitmq"

	osss "github.com/miuer/miu-netdisc/filestore/model/oss"
)

// Transfer -
func Transfer(writer *sql.DB, msg []byte) (err error) {

	tMeta := &rabbitmq.TransferMeta{}

	json.Unmarshal(msg, tMeta)

	err = osss.TransferToOss(tMeta)
	if err != nil {
		return err
	}

	err = mysql.UpdateFileMetaBySha1(
		writer,
		tMeta.FileName,
		tMeta.FileDestAddr,
		tMeta.FileSha1,
	)

	return err
}
