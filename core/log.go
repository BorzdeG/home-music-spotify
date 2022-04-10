package core

import log "github.com/sirupsen/logrus"

func InitLog() {
	log.SetLevel(log.TraceLevel)
	log.SetFormatter(
		&log.TextFormatter{
			FullTimestamp:          true,
			DisableLevelTruncation: true,
			PadLevelText:           true,
			ForceColors:            true,
		},
	)
	log.SetReportCaller(false)
}
