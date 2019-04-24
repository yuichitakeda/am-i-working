package scape

import (
	"time"
)

var holidays = map[int]map[time.Month][]int{
	2019: {
		time.April:     []int{19, 21}, // Sexta Santa, Tiradentes
		time.May:       []int{1},      // Trabalhador
		time.June:      []int{20, 21}, // Corpus Christi, Enforcado
		time.July:      []int{},
		time.August:    []int{15, 16},     // Adesão, Enforcado
		time.September: []int{7},          // Independência
		time.October:   []int{12, 14, 28}, // Crianças, Pós-Círio, Recírio
		time.November:  []int{2, 15},      // Finados, República
		time.December: []int{
			8, 23, 24, 25, 26, 27, 28, 29, 30, 31, // Conceição, Recesso
		},
	},
}

func GoalHours() time.Duration {
	belemTime := time.FixedZone("UTC-3", -3*60*60)
	now := time.Now().In(belemTime)
	year, month, today := now.Date()

	workDays := 0
	first := time.Date(year, month, 1, 12, 0, 0, 0, time.Local)
	for day := first; day.Month() == month; day = day.Add(24 * time.Hour) {
		switch day.Weekday() {
		case
			time.Tuesday,
			time.Wednesday,
			time.Thursday,
			time.Friday:
			workDays++
		}
		if day.Day() == today {
			break
		}
	}
	holidaysThisMonth := holidays[year][month]
	nHolidaysUntilToday := 0
	for i := range holidaysThisMonth {
		if holidaysThisMonth[i] > today {
			break
		}
		nHolidaysUntilToday++
	}
	goal := time.Duration(workDays-nHolidaysUntilToday) * 8 * time.Hour * 9 / 10
	return goal
}
