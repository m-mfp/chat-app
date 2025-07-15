package main

import (
	"fmt"
	"net/http"
)

func main() {
	ConnectDB()
	DB.AutoMigrate(&Message{})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Server is running")
	})

	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// fetch and return messages from DB
		} else if r.Method == http.MethodPost {
			// parse and create new message
		}
	})

	fmt.Println("Starting server on :8000...")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		panic(err)
	}
}
