package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	r := mux.NewRouter()

	r.HandleFunc("/users", createUser).Methods("POST")
	r.HandleFunc("/users", getAllUsers).Methods("GET")
	r.HandleFunc("/users/{id}", getUser).Methods("GET")
	r.HandleFunc("/users/{id}", updateUser).Methods("PUT")
	r.HandleFunc("/users/{id}", deleteUser).Methods("DELETE")

	log.Println("Starting server on :9000")
	log.Fatal(http.ListenAndServe(":9000", r))
}

func getDBConnection() (*sql.DB, error) {
	driver := os.Getenv("DB_DRIVER")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, name)
	return sql.Open(driver, dsn)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	db, err := getDBConnection()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		log.Printf("DB error: %v", err)
		return
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {

		}
	}(db)

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := CreateUser(db, user.Name, user.Email); err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		log.Printf("Create error: %v", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = fmt.Fprintln(w, "User created successfully")
	if err != nil {
		return
	}
}

func CreateUser(db *sql.DB, name, email string) error {
	query := "INSERT INTO users (name, email) VALUES (?, ?)"
	_, err := db.Exec(query, name, email)
	return err
}

func getAllUsers(w http.ResponseWriter, r *http.Request) {
	db, err := getDBConnection()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		log.Printf("DB error: %v", err)
		return
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			http.Error(w, "Database connection error", http.StatusInternalServerError)
		}
	}(db)

	users, err := GetAllUsers(db)
	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		log.Printf("Fetch error: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func GetAllUsers(db *sql.DB) ([]User, error) {
	rows, err := db.Query("SELECT id, name, email FROM users")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func getUser(w http.ResponseWriter, r *http.Request) {
	db, err := getDBConnection()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		log.Printf("DB error: %v", err)
		return
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			http.Error(w, "Database connection error", http.StatusInternalServerError)
		}
	}(db)

	idStr := mux.Vars(r)["id"]
	userId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := GetUser(db, userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error retrieving user", http.StatusInternalServerError)
			log.Printf("Get error: %v", err)
		}
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
		return nil, err
	}
	return &user, nil
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	db, err := getDBConnection()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		log.Printf("DB error: %v", err)
		return
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			http.Error(w, "Database connection error", http.StatusInternalServerError)
		}
	}(db)

	idStr := mux.Vars(r)["id"]
	userId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := UpdateUser(db, userId, user.Name, user.Email); err != nil {
		http.Error(w, "Error updating user", http.StatusInternalServerError)
		log.Printf("Update error: %v", err)
		return
	}

	fmt.Fprintln(w, "User updated successfully")
}

func UpdateUser(db *sql.DB, id int, name, email string) error {
	query := "UPDATE users SET name = ?, email = ? WHERE id = ?"
	_, err := db.Exec(query, name, email, id)
	return err
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	db, err := getDBConnection()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		log.Printf("DB error: %v", err)
		return
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			http.Error(w, "Database connection error", http.StatusInternalServerError)
		}
	}(db)

	idStr := mux.Vars(r)["id"]
	userId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := DeleteUser(db, userId); err != nil {
		http.Error(w, "Error deleting user", http.StatusInternalServerError)
		log.Printf("Delete error: %v", err)
		return
	}

	_, err = fmt.Fprintln(w, "User deleted successfully")
	if err != nil {
		return
	}
}

func DeleteUser(db *sql.DB, id int) error {
	query := "DELETE FROM users WHERE id = ?"
	_, err := db.Exec(query, id)
	return err
}
