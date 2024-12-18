package template

import (
	"fmt"
	"html/template"
	"time"
)

// Criar um novo FuncMap para ser usado nos templates
var TemplateFuncs = template.FuncMap{
	"formatTimeAgo": formatTimeAgo,
}

func formatTimeAgo(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "agora"
	case diff < time.Hour:
		minutes := int(diff.Minutes())
		return fmt.Sprintf("há %d min", minutes)
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		return fmt.Sprintf("há %d h", hours)
	case diff < 48*time.Hour:
		return "ontem"
	default:
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("há %d dias", days)
	}
}
