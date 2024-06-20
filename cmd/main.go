package main

import (
	"encoding/json"
	"fmt"
	"io"
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

func AddWordHandler(w http.ResponseWriter, r *http.Request) {
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
	// slog.Info("New dictionary record received - ", dict.Words)

	slog.Debug("Saving started")
	if err := saveDictionary(); err != nil {
		http.Error(w, "Error saving dictionary", http.StatusInternalServerError)
		return
	}
	slog.Debug("Saving complite")

	w.WriteHeader(http.StatusOK)
}

func DeleteWordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}

	var delete_word struct {
		Word string `json:"word"`
	}

	if err := json.NewDecoder(r.Body).Decode(&delete_word); err != nil {
		fmt.Println(r.Body)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	// slog.Info("Word for delete received - ", delete_word.Word)

	dict.mu.Lock()
	defer dict.mu.Unlock()

	delete(dict.Words, delete_word.Word)

	slog.Debug("Saving started")
	if err := saveDictionary(); err != nil {
		http.Error(w, "Error saving dictionary", http.StatusInternalServerError)
		return
	}
	slog.Debug("Saving complite")

	w.WriteHeader(http.StatusOK)

}

func GetWordsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	dict.mu.Lock()
	defer dict.mu.Unlock()

	json.NewEncoder(w).Encode(dict.Words)
	slog.Debug("The entire dictionary has been sent")
}

func GetWordsByLetterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	letter := r.URL.Query().Get("letter")
	if letter == "" {
		http.Error(w, "Letter query parameter is required", http.StatusBadRequest)
		return
	}
	// slog.Debug("Letter received - ", letter)

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
	slog.Info("Filter dictionary has been send")
}

func readDictionary() (map[string]string, error) {
	file, err := os.Open(dictionaryFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var result map[string]string
	err = json.Unmarshal(byteValue, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
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
	loadedDict, err := readDictionary()
	if err != nil {
		slog.Error("Error loading dictionary:", err)
	} else if loadedDict != nil {
		dict.Words = loadedDict
		slog.Info("Dictionary loaded successfully")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/add", AddWordHandler)
	mux.HandleFunc("/delete", DeleteWordHandler)
	mux.HandleFunc("/get", GetWordsHandler)
	mux.HandleFunc("/get-by-letter", GetWordsByLetterHandler)

	slog.Info("Starting server at port 8080")

	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		panic(err)
	}
}
