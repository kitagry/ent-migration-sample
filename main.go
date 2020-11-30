package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/kitagry/ent-migration-sample/ent"
	"github.com/kitagry/ent-migration-sample/ent/migrate"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	client, err := ent.Open("sqlite3", "./test.db?_fk=1")
	if err != nil {
		log.Fatalf("failed to open: %+v", err)
	}
	defer client.Close()

	// migration
	if err := client.Schema.Create(context.Background(), migrate.WithGlobalUniqueID(true)); err != nil {
		log.Fatalf("failed to migrate: %+v", err)
	}

	user, err := client.User.Query().First(context.Background())
	if err != nil {
		log.Fatalf("failed to get user: %+v", err)
	}
	if user == nil {
		user, err = client.User.Create().SetName("test user").Save(context.Background())
		if err != nil {
			log.Fatalf("failed to create user: %+v", err)
		}
	}

	http.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			todo, err := client.Todo.Create().SetTitle("hello").SetDescription("world").SetUser(user).Save(r.Context())
			if err != nil {
				log.Printf("failed to create todo: %+v", err)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("failed to create todo"))
				return
			}

			data, err := json.Marshal(todo)
			if err != nil {
				log.Printf("failed to marshal todo: %+v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to marshal todo"))
				return
			}
			w.Write(data)
		} else if r.Method == "GET" {
			todos, err := client.Todo.Query().WithUser().All(r.Context())
			if err != nil {
				log.Printf("failed to list todos: %+v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to list todos"))
				return
			}

			data, err := json.Marshal(todos)
			if err != nil {
				log.Printf("failed to marshal todos: %+v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("failed to marshal todos"))
				return
			}
			w.Write(data)
		}
	})
	log.Println(http.ListenAndServe(":8080", nil))
}
