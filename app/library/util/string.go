package util

import (
	"regexp"
	"time"
)

var numberRegex = regexp.MustCompile(`\D`)

func KeepNumbers(value string) string {
	return numberRegex.ReplaceAllString(value, "")
}

var MonthPTBR = map[time.Month]string{
	time.January:   "Janeiro",
	time.February:  "Fevereiro",
	time.March:     "Março",
	time.April:     "Abril",
	time.May:       "Maio",
	time.June:      "Junho",
	time.July:      "Julho",
	time.August:    "Agosto",
	time.September: "Setembro",
	time.October:   "Outubro",
	time.November:  "Novembro",
	time.December:  "Dezembro",
}
