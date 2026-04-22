package httpapi

import (
	"encoding/json"
	"net/http"

	"card-reading-system/internal/state"
)

func NewHandler(store *state.Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/card", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		card := store.Snapshot()
		payload := struct {
			UID *string `json:"uid"`
			ATR *string `json:"atr"`
		}{
			UID: emptyToNil(card.UID),
			ATR: emptyToNil(card.ATR),
		}

		_ = json.NewEncoder(w).Encode(payload)
	})

	return mux
}

func emptyToNil(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
