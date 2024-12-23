package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type CensorshipService struct {
	blockedWords []string
}

// RequestPayload представляет структуру запроса.
type RequestPayload struct {
	Comment string `json:"comment"`
}

// ResponsePayload представляет структуру ответа.
type ResponsePayload struct {
	Message string `json:"message"`
}

// NewCensorshipService создает новый экземпляр сервиса цензуры.
func NewCensorshipService(blockedWords []string) *CensorshipService {
	return &CensorshipService{blockedWords: blockedWords}
}

// validateComment проверяет комментарий на наличие запрещенных слов.
func (s *CensorshipService) validateComment(comment string) bool {
	for _, word := range s.blockedWords {
		if strings.Contains(strings.ToLower(comment), strings.ToLower(word)) {
			return false
		}
	}
	return true
}

// handler является обработчиком запросов.
func (s *CensorshipService) handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Генерация или извлечение сквозного идентификатора.
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = uuid.New().String()
	}

	log.Printf("[%s] Received request", requestID)

	// Проверяем метод.
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		log.Printf("[%s] Method not allowed", requestID)
		return
	}

	// Парсинг тела запроса.
	var payload RequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		log.Printf("[%s] Invalid request payload: %v", requestID, err)
		return
	}
	if payload.Comment == "" {
		http.Error(w, "Comment is empty", http.StatusBadRequest)
		log.Printf("[%s] Comment is empty", requestID)
		return
	}

	// Валидация комментария.
	if !s.validateComment(payload.Comment) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ResponsePayload{
			Message: "Comment contains inappropriate content",
		})
		log.Printf("[%s] Validation failed: %s", requestID, payload.Comment)
		return
	}

	// Возврат успешного ответа.
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ResponsePayload{
		Message: "Comment is valid",
	})
	log.Printf("[%s] Validation succeeded: %s (Processed in %s)", requestID, payload.Comment, time.Since(start))
}

func main() {
	blockedWords := []string{"spam", "offensive", "banned"}

	service := NewCensorshipService(blockedWords)
	http.HandleFunc("/validate", service.handler)

	log.Println("Censorship service running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
