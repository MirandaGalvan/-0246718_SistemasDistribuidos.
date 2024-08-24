package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type Record struct {
	Value  []byte `json:"value"`
	Offset uint64 `json:"offset"`
}

type Log struct {
	mu      sync.Mutex
	records []Record
}

var log Log

func main() {
	http.HandleFunc("/set", postRecord)
	http.HandleFunc("/get", getRecord)
	fmt.Println(" escuchando en http://localhost:3000")
	http.ListenAndServe(":3000", nil)
}

func postRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Record Record `json:"record"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.mu.Lock()
	req.Record.Offset = uint64(len(log.records))
	log.records = append(log.records, req.Record)
	log.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]uint64{"offset": req.Record.Offset})
}

func getRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Offset uint64 `json:"offset"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.mu.Lock()
	defer log.mu.Unlock()

	if req.Offset >= uint64(len(log.records)) {
		http.Error(w, "Registro no encontrado", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]Record{"record": log.records[req.Offset]})
}
