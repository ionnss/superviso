package docs

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Document struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func GetDocument(w http.ResponseWriter, r *http.Request) {
	docType := r.URL.Query().Get("type")

	docFiles := map[string]string{
		"terms":        "legal_docs/terms_of_service.md",
		"privacy":      "legal_docs/privacy_policy.md",
		"cancellation": "legal_docs/cancellation_policy.md",
		"fiscal":       "legal_docs/fiscal_guidelines.md",
		"contract":     "legal_docs/supervisor_contract.md",
	}

	if filePath, exists := docFiles[docType]; exists {
		content, err := os.ReadFile(filePath)
		if err != nil {
			http.Error(w, "Documento não encontrado", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(content))
	} else {
		http.Error(w, "Tipo de documento inválido", http.StatusBadRequest)
	}
}

func formatDocumentHTML(content, title string) string {
	lines := strings.Split(content, "\n")
	var html strings.Builder

	html.WriteString(fmt.Sprintf("<h1>%s</h1><div class=\"doc-text\">", title))

	for _, line := range lines {
		if matched, _ := regexp.MatchString(`^\d+\.`, line); matched {
			html.WriteString(fmt.Sprintf("<h2>%s</h2>", line))
		} else if matched, _ := regexp.MatchString(`^\s+[a-z]\)`, line); matched {
			html.WriteString(fmt.Sprintf("<p class=\"ms-4\">%s</p>", line))
		} else {
			html.WriteString(fmt.Sprintf("<p>%s</p>", line))
		}
	}

	html.WriteString("</div>")
	return html.String()
}

func getDocTitle(docType string) string {
	titles := map[string]string{
		"terms":        "Termos de Serviço",
		"privacy":      "Política de Privacidade",
		"cancellation": "Política de Cancelamento",
		"fiscal":       "Diretrizes Fiscais",
		"contract":     "Contrato de Supervisor",
	}
	return titles[docType]
}
