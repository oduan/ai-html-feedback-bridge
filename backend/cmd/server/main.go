package main

import (
	"log"
	"net/http"

	"github.com/oisin/ai-html-feedback-bridge/internal/config"
	"github.com/oisin/ai-html-feedback-bridge/internal/httpapi"
	"github.com/oisin/ai-html-feedback-bridge/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	db, err := storage.Open(cfg.SQLitePath)
	if err != nil {
		log.Fatalf("database error: %v", err)
	}
	defer db.Close()

	if err := storage.Migrate(db); err != nil {
		log.Fatalf("migration error: %v", err)
	}

	store := &storage.Store{DB: db}
	handler := httpapi.NewRouter(store, cfg.MaxContentSize)

	addr := ":" + cfg.PORT
	log.Printf("starting server on %s", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
