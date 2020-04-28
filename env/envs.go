package env

import (
	"log"

	"github.com/iterum-provenance/iterum-go/env"
	"github.com/iterum-provenance/iterum-go/util"
)

func init() {
	errIterum := env.VerifyIterumEnvs()
	errMinio := env.VerifyMinioEnvs()
	errDaemon := env.VerifyDaemonEnvs()
	errMessageq := env.VerifyMessageQueueEnvs()

	err := util.ReturnFirstErr(errIterum, errMinio, errDaemon, errMessageq)
	if err != nil {
		log.Fatalln(err)
	}
}
