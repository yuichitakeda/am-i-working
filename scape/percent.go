package scape

import (
	"time"
)

func Holidays(month time.Month, year int) int {
	switch month {
	case time.April:
		return 2
	}
	return 0
}

func GoalHours() time.Duration {
	belemTime := time.FixedZone("UTC-3", -3*60*60)
	now := time.Now().In(belemTime)
	month := now.Month()
	year := now.Year()
	workDays := 0
	first := time.Date(year, month, 1, 12, 0, 0, 0, time.Local)
	for day := first; day.Month() == month; day = day.Add(24 * time.Hour) {
		switch day.Weekday() {
		case time.Sunday, time.Monday, time.Saturday:

		default:
			workDays++
		}
		if day.Day() == now.Day() {
			break
		}
	}
	holy := Holidays(month, year)
	goal := time.Duration(workDays-holy) * 8 * time.Hour * 9 / 10
	return goal
}
