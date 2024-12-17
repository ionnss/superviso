package utils

import (
	"strings"
	"time"
)

// FormatDate formata data para exibição (DD/MM/YYYY)
func FormatDate(t time.Time) string {
	return t.Format("02/01/2006")
}

// FormatTime formata hora para exibição (HH:MM)
func FormatTime(t string) string {
	if t == "" {
		return ""
	}
	timeObj, _ := time.Parse("15:04:05", t)
	return timeObj.Format("15:04")
}

// FormatWeekday retorna o nome do dia da semana
func FormatWeekday(day int) string {
	weekdays := map[int]string{
		0: "Domingo",
		1: "Segunda",
		2: "Terça",
		3: "Quarta",
		4: "Quinta",
		5: "Sexta",
		6: "Sábado",
	}
	return weekdays[day]
}

// ParseWeekDays converte string de dias em slice
func ParseWeekDays(days string) []string {
	if days == "" {
		return []string{}
	}
	return strings.Split(days, ",")
}
