package main

import (
	"database/sql"
	"flag"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const (
	dbFile   = "users.db"
	table    = "auth"
	coluser = "user"
	colpass = "password"
	colmail = "email"
)

func main() {
	// Connect to the database
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	defer db.Close()

	// Create the auth table if it doesn't exist
	createTableQuery := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		%s TEXT PRIMARY KEY,
		%s TEXT NOT NULL,
		%s TEXT NOT NULL,
    is_admin BOOL
	);`, table, coluser, colpass, colmail)
	_, err = db.Exec(createTableQuery)
	if err != nil {
		fmt.Println("Error creating table:", err)
		return
	}

	action := flag.String("action", "", "Action to use (add/delete)")
	username := flag.String("user", "", "Username to use")
	password := flag.String("pass", "", "Password to use")
	email := flag.String("email", "", "email to use (optional)")
	is_admin := flag.Bool("admin", false, "Set as admin (optional)")
	showusers := flag.Bool("show", false, "Show Users")
	flag.Parse()

	if *showusers == true {
		showUsers(db)
	}
	if *action != "" {
		switch *action {
		case "add":
			err = addUser(db, *username, *password, *email, *is_admin)
			if err != nil {
				fmt.Println("Error adding user:", err)
				return
			}
		case "delete":
			err = deleteUser(db, *username)
			if err != nil {
				fmt.Println("Error deleting user:", err)
				return
			}
		}
	}
	// default:
	// 	fmt.Println("Invalid action.")
	// 	fmt.Println("Usage: ./cmd <action> <username> <password>")
	// 	fmt.Println("Example: ./cmd add user1 password1")
	// 	fmt.Println("Example: ./cmd delete user1")
	// 	return
	// }
}

func showUsers(db *sql.DB) error {
	rows, err := db.Query("SELECT * FROM auth")
	if err != nil {
		return err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	values := make([]interface{}, len(columns))
	valuePointers := make([]interface{}, len(columns))
	for i := range columns {
		valuePointers[i] = &values[i]
	}
	for rows.Next() {
		if err := rows.Scan(valuePointers...); err != nil {
			return err
		}
		for i, column := range columns {
			fmt.Printf("%s: %s\n", column, values[i])
		}
		fmt.Println()
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}

// addUser adds a user to the auth table
func addUser(db *sql.DB, username, password, email string, is_admin bool) error {
	// Check if the user already exists
	checkQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s = ?;", table, coluser)
	var count int
	err := db.QueryRow(checkQuery, username).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		fmt.Println("User already exists")
		return nil
	}
	// Encrypt password with bcrypt
	hashedpassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil{
		fmt.Printf("Error %s ", err)
		return err
	}
	// Insert the new user
	insertQuery := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, is_admin) VALUES (?, ?, ?, ?);", table, coluser, colpass, colmail)
	_, err = db.Exec(insertQuery, username, hashedpassword, email, is_admin)
	if err != nil {
		return err
	}
	fmt.Println("User added successfully")
	return nil
}

// deleteUser deletes a user from the auth table
func deleteUser(db *sql.DB, username string) error {
	deleteQuery := fmt.Sprintf("DELETE FROM %s WHERE user = ?;", table)
	_, err := db.Exec(deleteQuery, username)
	if err != nil {
		return err
	}
	fmt.Println("User deleted successfully")
	return nil
}
