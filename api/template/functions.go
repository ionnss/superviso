package template

import (
	"fmt"
	"html/template"
	"time"
)

var TemplateFuncs = template.FuncMap{
	"formatDate":    formatDate,
	"formatTime":    formatTime,
	"formatWeekday": formatWeekday,
	"dict": func(values ...interface{}) (map[string]interface{}, error) {
		if len(values)%2 != 0 {
			return nil, fmt.Errorf("invalid dict call")
		}
		dict := make(map[string]interface{}, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				return nil, fmt.Errorf("dict keys must be strings")
			}
			dict[key] = values[i+1]
		}
		return dict, nil
	},
}

func formatDate(t time.Time) string {
	return t.Format("02/01/2006")
}

func formatTime(t time.Time) string {
	return t.Format("15:04")
}

func formatWeekday(t time.Time) string {
	weekday := t.Weekday()
	switch weekday {
	case time.Monday:
		return "Segunda-feira"
	case time.Tuesday:
		return "Terça-feira"
	case time.Wednesday:
		return "Quarta-feira"
	case time.Thursday:
		return "Quinta-feira"
	case time.Friday:
		return "Sexta-feira"
	case time.Saturday:
		return "Sábado"
	default:
		return "Domingo"
	}
}
