package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

const (
	db_driver   = "mysql"
	db_user     = "root"
	db_password = "root"
	db_name     = "golang_crud"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/users", createUser).Methods("POST")
	r.HandleFunc("/users/{id}", getUser).Methods("GET")
	r.HandleFunc("/users/{id}", updateUser).Methods("PUT")
	r.HandleFunc("/users/{id}", deleteUser).Methods("DELETE")

	log.Println("Starting server on :9000")
	if err := http.ListenAndServe(":9000", r); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}

}

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func createUser(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open(db_driver, db_user+":"+db_password+"@/"+db_name)

	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		log.Printf("Error connecting to database: %v", err)
		return
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			http.Error(w, "Database connection error", http.StatusInternalServerError)
		}
	}(db)

	var user User
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		return
	}

	err = CreateUser(db, user.Name, user.Email)

	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		log.Printf("Error creating user: %v", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
    fmt.Fprintln(w, "User created successfully")

}

func CreateUser(db *sql.DB, name, email string) error {
	query := "INSERT INTO users (name, email) VALUES (?, ?)"
	_, err := db.Exec(query, name, email)

	if err != nil {
		return fmt.Errorf("error inserting user: %w", err)
	}

	return nil
}

func getUser(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open(db_driver, db_user+":"+db_password+"@/"+db_name)

	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		log.Printf("Error connecting to database: %v", err)
		return
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			http.Error(w, "Database connection error", http.StatusInternalServerError)
		}
	}(db)

	//get id parameter from the request
	vars := mux.Vars(r)
	idStr := vars["id"]

	userId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		log.Printf("Error converting user ID: %v", err)
		return
	}

	user, err := GetUser(db, userId)
	if err != nil {
		http.Error(w, "Error retrieving user", http.StatusInternalServerError)
		log.Printf("Error retrieving user: %v", err)
		return
	}

	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		return
	}
}

func GetUser(db *sql.DB, id int) (*User, error) {
	query := "SELECT id, name, email FROM users WHERE id = ?"
	row := db.QueryRow(query, id)

	var user User
	err := row.Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("error retrieving user: %w", err)
	}

	return &user, nil
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open(db_driver, db_user+":"+db_password+"@/"+db_name)

	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		log.Printf("Error connecting to database: %v", err)
		return
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			http.Error(w, "Database connection error", http.StatusInternalServerError)
		}
	}(db)

	vars := mux.Vars(r)
	idStr := vars["id"]
	userId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		log.Printf("Error converting user ID: %v", err)
		return
	}

	var user User

	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		log.Printf("Error decoding request body: %v", err)
		return
	}

	err = UpdateUser(db, userId, user.Name, user.Email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	_, err = fmt.Fprint(w, "User updated successfully")
	if err != nil {
		return
	}
}

func UpdateUser(db *sql.DB, id int, name, email string) error {
	query := "UPDATE users SET name = ?, email = ? WHERE id = ?"
	_, err := db.Exec(query, name, email, id)

	if err != nil {
		return err
	}

	return nil
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open(db_driver, db_user+":"+db_password+"@/"+db_name)

	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		log.Printf("Error connecting to database: %v", err)
		return
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			http.Error(w, "Database connection error", http.StatusInternalServerError)
		}
	}(db)

	vars := mux.Vars(r)
	idStr := vars["id"]
	userId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		log.Printf("Error converting user ID: %v", err)
		return
	}

	user := DeleteUser(db, userId)
	if user == nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	_, err = fmt.Fprintln(w, "User deleted successfully")
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		return
	}
}

func DeleteUser(db *sql.DB, id int) error {
	query := "DELETE FROM users WHERE id = ?"
	_, err := db.Exec(query, id)
	if err != nil {
		return err
	}
	return nil
}
