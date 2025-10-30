package logger

import (
	"os"

	"github.com/sirupsen/logrus"

)

func Init() {
	logrus.SetOutput(os.Stdout)
	logrus.SetReportCaller(true)

	logrus.SetFormatter(&Formatter{
		TimestampFormat: "2006-01-02 15:04:05.000",
		NoColors:        true,
		NoFieldsColors:  true,
	})
}
