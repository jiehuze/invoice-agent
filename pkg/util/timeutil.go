package util

import (
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	TIME_LAYOUT            = "2006-01-02 15:04:05"
	LAYOUT_TIME_STAMP      = "2006-01-02 15:04:05"
	LAYOUT_TIME_STAMPMILLI = "2006-01-02 15:04:05.000"
	LAYOUT_TIME_STAMPDAY   = "2006-01-02"
	LAYOUT_TIME_BLENDDAY   = "20060102"
	LAYOUT_TIME_BLENDSEC   = "20060102150405"
)

func CurrentTimeFormat() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func DatetimeToUnix(date string) int64 {
	t, _ := time.Parse(TIME_LAYOUT, date)
	tLocal := t.Local()
	return tLocal.Unix()
}

func DatetimeToTime(date string) time.Time {
	t, _ := time.Parse(TIME_LAYOUT, date)
	tLocal := t.Local()

	return tLocal
}

func UnixToString(timestamp int64) string {
	tm := time.Unix(timestamp, 0)
	return tm.Format(TIME_LAYOUT)
}

func UnixToTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

func GetDayAddMonth(dt string, month int) string {
	if len(dt) != 8 {
		log.Infof("Format for dt is error; dt: %s", dt)
		return ""
	}

	const TIME_LAYOUT = "20060102"
	var lastT time.Time
	t, _ := time.Parse(TIME_LAYOUT, dt)
	log.Info("time :", t)

	if t.Month() == 3 {
		var remain int = 0
		if IsLeapYear(t.Year()) && t.Day() > 29 {
			remain = 29 - t.Day()
		} else if t.Day() >= 29 {
			remain = 28 - t.Day()
		}

		lastT = t.AddDate(0, month, remain)
	} else if t.Day() == 31 {
		lastT = t.AddDate(0, month, -1)
	} else {
		lastT = t.AddDate(0, -1, 0)
	}

	log.Info("lastT :", lastT)
	return lastT.Format(TIME_LAYOUT)
}

func IsLeapYear(year int) bool {
	if year%4 == 0 && year%100 != 0 || year%400 == 0 {
		return true
	}
	return false
}
