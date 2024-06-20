package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetWordsHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/get", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetWordsHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestAddWordHandler(t *testing.T) {
	tests := []struct {
		method         string
		body           interface{}
		expectedStatus int
	}{
		{
			method: http.MethodPost,
			body: struct {
				Word       string `json:"word"`
				Definition string `json:"definition"`
			}{
				Word:       "test",
				Definition: "This is a test definition",
			},
			expectedStatus: http.StatusOK,
		},
		{
			method:         http.MethodGet,
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			method:         http.MethodPost,
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		var bodyBytes []byte
		if test.body != nil {
			bodyBytes, _ = json.Marshal(test.body)
		}

		req, err := http.NewRequest(test.method, "/add", bytes.NewBuffer(bodyBytes))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(AddWordHandler)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != test.expectedStatus {
			t.Errorf("Handler returned wrong status code: got %v want %v", status, test.expectedStatus)
		}
	}

	if dict.Words["test"] != "This is a test definition" {
		t.Errorf("Word not added to dictionary: got %v", dict.Words["test"])
	}
}

func TestGetWordsByLetterHandler(t *testing.T) {
	tests := []struct {
		method         string
		letter         string
		expectedStatus int
		expectedWords  map[string]string
	}{
		{
			method:         http.MethodGet,
			letter:         "t",
			expectedStatus: http.StatusOK,
			expectedWords: map[string]string{
				"test": "This is a test definition",
			},
		},
		{
			method:         http.MethodGet,
			letter:         "c",
			expectedStatus: http.StatusOK,
			expectedWords:  map[string]string{},
		},
		{
			method:         http.MethodPost,
			letter:         "a",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedWords:  nil,
		},
		{
			method:         http.MethodGet,
			letter:         "",
			expectedStatus: http.StatusBadRequest,
			expectedWords:  nil,
		},
	}

	for _, test := range tests {
		req, err := http.NewRequest(test.method, "/get-by-letter", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		q := req.URL.Query()
		q.Add("letter", test.letter)
		req.URL.RawQuery = q.Encode()

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(GetWordsByLetterHandler)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != test.expectedStatus {
			t.Errorf("Handler returned wrong status code: got %v want %v", status, test.expectedStatus)
		}

		if test.expectedStatus == http.StatusOK {
			var responseWords map[string]string
			if err := json.NewDecoder(rr.Body).Decode(&responseWords); err != nil {
				t.Errorf("Failed to decode response: %v", err)
			}

			for word, definition := range test.expectedWords {
				if responseWords[word] != definition {
					t.Errorf("Incorrect correspondence between letter and definition.")
				}
			}
		}
	}
}

func TestDeleteWordHandler(t *testing.T) {
	tests := []struct {
		method         string
		body           interface{}
		expectedStatus int
	}{
		{
			method:         http.MethodGet,
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			method: http.MethodPost,
			body: struct {
				Word string `json:"word"`
			}{
				Word: "test",
			},
			expectedStatus: http.StatusOK,
		},
		{
			method:         http.MethodPost,
			body:           "invalid json",
			expectedStatus: http.StatusBadRequest,
		}}

	for _, test := range tests {
		var bodyBytes []byte
		if test.body != nil {
			bodyBytes, _ = json.Marshal(test.body)
		}

		req, err := http.NewRequest(test.method, "/delete", bytes.NewBuffer(bodyBytes))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		rr := httptest.NewRecorder()

		handler := http.HandlerFunc(DeleteWordHandler)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != test.expectedStatus {
			t.Errorf("Handler returned wrong status code: got %v want %v", status, test.expectedStatus)
		}
	}
}
