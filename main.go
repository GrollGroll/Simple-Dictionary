package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
)

const dictionaryFile = "dictionary.json"

type Dictionary struct {
	Words map[string]string
	mu    sync.Mutex
}

var dict = Dictionary{
	Words: make(map[string]string),
}

func addWordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Word       string `json:"word"`
		Definition string `json:"definition"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	dict.mu.Lock()
	defer dict.mu.Unlock()

	dict.Words[req.Word] = req.Definition
	slog.Debug("Got new record", dict.Words)

	slog.Debug("Saving started")
	if err := saveDictionary(); err != nil {
		http.Error(w, "Error saving dictionary", http.StatusInternalServerError)
		return
	}
	slog.Debug("Saving complite")

	w.WriteHeader(http.StatusOK)
}

func getWordsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	dict.mu.Lock()
	defer dict.mu.Unlock()

	json.NewEncoder(w).Encode(dict.Words)
	slog.Debug("Whole dictionary was send")
}

func getWordsByLetterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	letter := r.URL.Query().Get("letter")
	if letter == "" {
		http.Error(w, "Letter query parameter is required", http.StatusBadRequest)
		return
	}
	slog.Debug("Got letter", letter)

	letter = strings.ToLower(letter)

	dict.mu.Lock()
	defer dict.mu.Unlock()

	filteredWords := make(map[string]string)
	for word, definition := range dict.Words {
		if strings.HasPrefix(strings.ToLower(word), letter) {
			filteredWords[word] = definition
		}
	}

	json.NewEncoder(w).Encode(filteredWords)
	slog.Info("Filter dictionary was send")
}

func loadDictionary() error {
	dict.mu.Lock()
	defer dict.mu.Unlock()

	file, err := os.Open(dictionaryFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // файл не существует, это не ошибка
		}
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(&dict.Words)
}

func saveDictionary() error {
	file, err := os.OpenFile(dictionaryFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(dict.Words)
}

func main() {
	if err := loadDictionary(); err != nil {
		slog.Error("Error loading dictionary:", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/add", addWordHandler)
	mux.HandleFunc("/get", getWordsHandler)
	mux.HandleFunc("/get-by-letter", getWordsByLetterHandler)

	slog.Info("Starting server at port 8080")

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
