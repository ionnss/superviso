package docs

import (
	"encoding/json"
	"net/http"
	"os"
)

type Document struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func GetDocument(w http.ResponseWriter, r *http.Request) {
	docType := r.URL.Query().Get("type")

	// Mapa de arquivos
	docFiles := map[string]string{
		"terms":        "legal_docs/terms_of_service.txt",
		"privacy":      "legal_docs/privacy_policy.txt",
		"cancellation": "legal_docs/cancellation_policy.txt",
		"fiscal":       "legal_docs/fiscal_guidelines.txt",
		"contract":     "legal_docs/supervisor_contract.txt",
	}

	if filePath, exists := docFiles[docType]; exists {
		content, err := os.ReadFile(filePath)
		if err != nil {
			http.Error(w, "Documento não encontrado", http.StatusNotFound)
			return
		}

		doc := Document{
			Title:   getDocTitle(docType),
			Content: string(content),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(doc)
	} else {
		http.Error(w, "Tipo de documento inválido", http.StatusBadRequest)
	}
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
