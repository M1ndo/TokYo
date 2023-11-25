// Date: 2023/07/25
// Middleware for authentication/Logging.
package app

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type Middleware struct {
	AppInstance *App
	DB          *sql.DB
}

// Initialize the SQLite database connection
func (m *Middleware) InitializeDB() error {
	db, err := sql.Open("sqlite3", "users.db") // Replace "users.db" with your desired database file name
	if err != nil {
		return err
	}
	m.DB = db
	// Create the "auth" table if it doesn't exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS auth (
		user TEXT PRIMARY KEY,
		password TEXT,
		email TEXT,
    is_admin BOOL
	)`)
	if err != nil {
		return err
	}
	return nil
}

// Check if user exists.
func (m *Middleware) finduser(username string, db *sql.DB) (bool, error) {
	query := "SELECT user from AUTH where user = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		return false, err
	}
	defer stmt.Close()
	var user string
	err = stmt.QueryRow(username).Scan(&user)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (m *Middleware) authenticate(username, password string, db *sql.DB) bool {
	// Check if the user exists
	query := "SELECT password FROM auth WHERE user = ?"
	stmt, err := db.Prepare(query)
	if err != nil {
		m.AppInstance.Logger.Log.Error("[ERROR] Failed to prepare query:", err)
		return false
	}
	defer stmt.Close()
	var hashedPassword string
	err = stmt.QueryRow(username).Scan(&hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		log.Println("[ERROR] Failed to execute query:", err)
		return false
	}
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false
		}
		log.Println("[ERROR] Failed to compare passwords:", err)
		return false
	}
	return true
}

// Main Middleware
func (m *Middleware) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.AppInstance.Logger.Log.Println("[+] Middleware has been executed.")
		session, err := m.AppInstance.Sessions.Get(r, "session")
		if err != nil {
			m.AppInstance.FailedSession(w, r)
			return
		}
		if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
			m.AppInstance.Deniedhandler(w, r, &ErrorHandler{Error: "Access Denied"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Check if user is admin
func (m *Middleware) is_admin(username string) bool {
	query := "SELECT is_admin from auth where user = ?"
	ftmt, err := m.DB.Prepare(query)
	if err != nil {
		return false
	}
	defer ftmt.Close()
	var isAdmin bool
	err = ftmt.QueryRow(username).Scan(&isAdmin)
	if err != nil {
		return false
	}
	return isAdmin
}

// Function to authenticate a user and set the session
func (m *Middleware) AuthenticateAndSetSession(username, password string, w http.ResponseWriter, r *http.Request) bool {
	if m.DB == nil {
		m.AppInstance.Logger.Log.Println("[ERROR] Database connection is not initialized.")
		return false
	}
	if !m.authenticate(username, password, m.DB) {
		return false
	}
	isAdmin := m.is_admin(username)
	sessionID, err := uuid.NewUUID()
	if err != nil {
		m.AppInstance.Logger.Log.Println("[ERROR] Failed to generate session ID:", err)
		return false
	}
	session, err := m.AppInstance.Sessions.Get(r, "session")
	if err != nil {
		m.AppInstance.Logger.Log.Println("[ERROR] Failed to get session:", err)
		return false
	}
	session.Values["username"] = username
	session.Values["sessionID"] = sessionID.String()
	session.Values["authenticated"] = true
	session.Values["restrictedAcc"] = isAdmin
	err = session.Save(r, w)
	if err != nil {
		m.AppInstance.Logger.Log.Println("[ERROR] Failed to save session:", err)
		return false
	}
	return true
}

// Sign up and set set a session
func (m *Middleware) SignUpAndSetSession(name, password, email string, w http.ResponseWriter, r *http.Request) (bool, error) {
	if m.DB == nil {
		m.AppInstance.Logger.Log.Println("[ERROR] Database connection is not initialized.")
		return false, errors.New("Database seem to be offline")
	}
	isuser, err := m.finduser(name, m.DB)
	if err != nil {
		m.AppInstance.Logger.Log.Error(err)
		return false, err
	}
	if !isuser {
		query := "INSERT INTO auth (user, password, email) values (?, ?, ?)"
		if len(password) > 72 {
			return false, errors.New("Password is too long")
		}
		hashedpass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return false, err
		}
		_, err = m.DB.Exec(query, name, hashedpass, email)
		if err != nil {
			return false, err
		}
		isconnect := m.AuthenticateAndSetSession(name, password, w, r)
		if isconnect {
			return true, nil
		} else {
			return false, errors.New("Cannot authenticated")
		}
	} else {
		return false, errors.New("User already exists!")
	}
}

// Log Visitor Ip And User-Agent (Can be disabled on the cmd options).
func (m *Middleware) LogUser(r *http.Request) {
	if m.AppInstance.Config.Logging {
		ip := r.Header.Get("CF-Connecting-IP")
		if ip == "" {
			ip = r.Header.Get("X-Forwarded-For")
		}
		if ip == "" {
			ip = "0xdeadbeef"
		}
		userAgent := r.UserAgent()
		m.AppInstance.Logger.Log.Logf("Connected Ip %s, User-agent %s", ip, userAgent)
	}
}
