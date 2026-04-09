package main

import (
	"log"
	"net/http"
	"os"

	"github.com/loic-vermeire/dx-connect-ci-scaffold/services/api/internal/handler"
	"github.com/loic-vermeire/dx-connect-ci-scaffold/services/api/internal/store"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	items := store.NewItemStore()
	h := handler.New(items)
	r := handler.NewRouter(h)

	log.Printf("listening on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
