// Package app manages main application server
// Created By ybenel
// Date: 26/06/2019

package app

import (
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	fsnotify "github.com/fsnotify/fsnotify"
	feeds "github.com/gorilla/feeds"
	mux "github.com/gorilla/mux"
	sessions "github.com/gorilla/sessions"
	secure "github.com/unrolled/secure"
	mylog "gitlab.com/DaFunk1/TokYo/pkg/log"
	media "gitlab.com/DaFunk1/TokYo/pkg/media"
	onionkey "gitlab.com/DaFunk1/TokYo/pkg/onionkey"
	// "github.com/sirupsen/logrus"
	// errlog "github.com/snwfdhmp/errlog"
)

// App represents main application.
type App struct {
	Config    *Config
	Middleware *Middleware
	Library   *media.Library
	Watcher   *fsnotify.Watcher
	Templates *template.Template
	Tor       *tor
	Listener  net.Listener
	Router    *mux.Router
	Sessions  *sessions.CookieStore
	Mdata     MData
	Logger    *mylog.Logger
	Debug     *DebugConfig
}

// NewApp returns a new instance of App from Config.
func NewApp(cfg *Config) (*App, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	a := &App{
		Config: cfg,
		Middleware: &Middleware{},
		Logger: &mylog.Logger{},
	}
	log, err := a.Logger.NewLogger() 
	if err != nil {
		return nil, err
	}
	// Setup Library
	a.Library = media.NewLibrary()
	// Setup Logger
	a.Logger = log
	// Setup debugger
	a.Debug = &DebugConfig{}
	a.Debug.Config = a.Debug.DefaultConfig()
	// a.Debug.Config.PrintFunc = logrus.Errorf 
	// a.Debug.Config.PrintFunc = a.Logger.Log.Fatalf 
	a.Debug.Logger = a.Debug.NewDebug() 
	// Setup Watcher
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	a.Watcher = w
	// Setup Listener
	ln, err := newListener(cfg.Server)
	if err != nil {
		return nil, err
	}
	a.Listener = ln
	// Setup Templates
	a.Templates = template.Must(template.ParseGlob("templates/*"))
	// Setup Tor
	if cfg.Tor.Enable {
		t, err := newTor(cfg.Tor)
		if err != nil {
			return nil, err
		}
		a.Tor = t
	}
	// Setup Sessions
	store := sessions.NewCookieStore([]byte("SOMETHING"))
	// store := sessions.NewCookieStore([]byte(os.Getenv("SEC_SESSION")))
	a.Sessions = store
	a.Middleware.AppInstance = a
	// Initialize DB
	a.Middleware.InitializeDB()
	// Setup Router
	secureMiddleware := secure.New(secure.Options{
		FrameDeny: true,
		// AllowedHosts:          []string{"ybenel\\.cf", "test\\.ybenel\\.cf", "192.\\168\\.1\\.110:9110"},
		// AllowedHostsAreRegex:  true,
		// HostsProxyHeaders:     []string{"X-Forwarded-Host"},
		// SSLRedirect:           false,
		// SSLHost:               "stream.ybenel.cf",
		SSLProxyHeaders:       map[string]string{"X-Forwarded-Proto": "https"},
		STSSeconds:            31536000,
		STSIncludeSubdomains:  true,
		STSPreload:            true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "script-src 'self' 'unsafe-inline'",
	})
	r := mux.NewRouter().StrictSlash(true)
	r.Use(secureMiddleware.Handler)
	// authRoute := r.NewRoute().Subrouter()
	// authRoute.Use(a.Middleware.AuthMiddleware)
	// authRoute.HandleFunc("/v/{id}.{ext}", a.videoHandler).Methods("GET")
	// authRoute.HandleFunc("/v/{prefix}/{id}.{ext}", a.videoHandler).Methods("GET")
	// authRoute.HandleFunc("/v/{id}.{ext}", a.videoHandler).Methods("GET")
	// authRoute.HandleFunc("/v/{id}", a.pageHandler).Methods("GET")
	// authRoute.HandleFunc("/v/{prefix}/{id}", a.pageHandler).Methods("GET")
	r.HandleFunc("/", a.indexHandler).Methods("GET")
	r.HandleFunc("/t/{id}", a.thumbHandler).Methods("GET")
	r.HandleFunc("/t/{prefix}/{id}", a.thumbHandler).Methods("GET")
	r.HandleFunc("/feed.xml", a.rssHandler).Methods("GET")
	r.HandleFunc("/login", a.loginHandler).Methods("GET", "POST")
	r.HandleFunc("/logout", a.logoutHandler).Methods("GET", "POST")
	r.HandleFunc("/signup", a.signupHandler).Methods("GET", "POST")
	r.HandleFunc("/v/{id}.{ext}", a.videoHandler).Methods("GET")
	r.HandleFunc("/v/{prefix}/{id}.{ext}", a.videoHandler).Methods("GET")
	r.HandleFunc("/v/{id}.{ext}", a.videoHandler).Methods("GET")
	r.HandleFunc("/v/{id}", a.pageHandler).Methods("GET")
	r.HandleFunc("/v/{prefix}/{id}", a.pageHandler).Methods("GET")
	r.NotFoundHandler = http.HandlerFunc(a.Unfoundhandler) // Set custom 404 handler
	// Static file handler
	fsHandler := http.StripPrefix(
		"/static/",
		http.FileServer(http.Dir("./static/")),
	)
	r.PathPrefix("/static/").Handler(fsHandler).Methods("GET")
	a.Router = r
	return a, nil
}

// Run imports the library and starts server.
func (a *App) Run() error {
	if a.Tor != nil {
		var err error
		cs := a.Config.Server
		key := a.Tor.OnionKey
		if key == nil {
			key, err = onionkey.GenerateKey()
			if err != nil {
				return err
			}
		}
		onion, err := key.Onion()
		if err != nil {
			return err
		}
		onion.Ports[80] = fmt.Sprintf("%s:%d", cs.Host, cs.Port)
		err = a.Tor.Controller.AddOnion(onion)
		if err != nil {
			return err
		}
		a.Logger.Log.Infof("Onion service: http://%s.onion", onion.ServiceID)
	}
	for _, pc := range a.Config.Library {
		p := &media.Path{
			Path:   pc.Path,
			Prefix: pc.Prefix,
			Private: pc.Private,
		}
		err := a.Library.AddPath(p)
		if err != nil {
			return err
		}
		err = a.Library.Import(a.Logger, p)
		if err != nil {
			return err
		}
		// a.Logger.Log.Printf(p.Path) // Prints Video Paths
		a.Watcher.Add(p.Path)
	}
	go startWatcher(a)

	return http.Serve(a.Listener, a.Router)
}

// Get Prefix
func getPrefix(id string) string {
	parts := strings.Split(id, "/")
	if len(parts) > 1 {
		return parts[0]
	}
	return "General"
}

// HTTP handler for /
func (a *App) indexHandler(w http.ResponseWriter, r *http.Request) {
	a.Logger.Log.Info("/")
	a.Logger.Log.Infof("Visitor IP: %s, User-Agent: %s", r.RemoteAddr, r.UserAgent())
	pl := a.Library.Playlist()
	sections := make(map[string][]*media.Video)
	for _, video := range pl {
		prefix := getPrefix(video.ID)
		sections[prefix] = append(sections[prefix], video)
	}
	session, err := a.Sessions.Get(r, "session")
	if err != nil {
		a.FailedSession(w, r)
	}
	authenticated, ok := session.Values["authenticated"].(bool)
	if !ok {
		authenticated = false
	}
	a.Templates.ExecuteTemplate(w, "index.html", &struct {
		Playlist media.Playlist
		Sections map[string][]*media.Video
		Auth bool
	}{
		Sections: sections,
		Playlist: pl,
		Auth: authenticated,
	})
}

// HTTP handler for /v/id
func (a *App) pageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	prefix, ok := vars["prefix"]
	if ok {
		id = path.Join(prefix, id)
	}
	a.Logger.Log.Infof("/v/%s", id)
	a.Logger.Log.Infof("Visitor IP: %s, User-Agent: %s", r.RemoteAddr, r.UserAgent())
	playing, ok := a.Library.Videos[id]
	if !ok {
		a.Unfoundhandler(w, r)
		return
	}
  auth, err := a.MediaAcess(w, r, playing)
	if err != nil {
		a.Deniedhandler(w, r, &ErrorHandler{Error: err.Error()})
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	a.Templates.ExecuteTemplate(w, "video.html", &struct {
		Playing  *media.Video
		FileType template.HTML
		Playlist media.Playlist
		Auth bool
	}{
		Playing:  playing,
		FileType: template.HTML(playing.FileType),
		Playlist: a.Library.Playlist(),
		Auth:     auth,
	})
}

// HTTP handler for video streaming
func (a *App) videoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	prefix, ok := vars["prefix"]
	if ok {
		id = path.Join(prefix, id)
	}
	ext := vars["ext"]
	m, ok := a.Library.Videos[id]
	if !ok {
		a.Logger.Log.Errorf("Not Ok! File %s.%s not found", id, ext)
		http.NotFound(w, r)
		return
	}
  _, err := a.MediaAcess(w, r, m)
	if err != nil {
		a.Deniedhandler(w, r, &ErrorHandler{Error: err.Error()})
	}
	contentType := a.Library.GetContentType(ext)
	filename := m.Title + "." + ext
	disposition := "attachment; filename=\"" + filename + "\""
	w.Header().Set("Content-Disposition", disposition)
	w.Header().Set("Content-Type", contentType)
	http.ServeFile(w, r, m.Path)
}

// HTTP handler for /t/id
func (a *App) thumbHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// fmt.Print(vars)
	id := vars["id"]
	prefix, ok := vars["prefix"]
	if ok {
		id = path.Join(prefix, id)
	}
	a.Logger.Log.Infof("/t/%s", id)
	m, ok := a.Library.Videos[id]
	if !ok {
		a.Logger.Log.Warn("Not Ok! File %s.%s not found", id)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=7776000")
	if m.ThumbType == "" {
		w.Header().Set("Content-Type", "image/jpeg")
		http.ServeFile(w, r, "static/thumb.jpg")
	} else {
		w.Header().Set("Content-Type", m.ThumbType)
		w.Write(m.Thumb)
	}
}

// HTTP handler for /feed.xml
func (a *App) rssHandler(w http.ResponseWriter, r *http.Request) {
	cfg := a.Config.Feed
	now := time.Now()
	f := &feeds.Feed{
		Title:       cfg.Title,
		Link:        &feeds.Link{Href: cfg.Link},
		Description: cfg.Description,
		Author: &feeds.Author{
			Name:  cfg.Author.Name,
			Email: cfg.Author.Email,
		},
		Created:   now,
		Copyright: cfg.Copyright,
	}
	var externalURL string
	if len(cfg.ExternalURL) > 0 {
		externalURL = cfg.ExternalURL
	} else {
		host := a.Config.Server.Host
		port := a.Config.Server.Port
		externalURL = fmt.Sprintf("http://%s:%d", host, port)
	}
	for _, v := range a.Library.Playlist() {
		u, err := url.Parse(externalURL)
		if err != nil {
			a.Logger.Log.Warn(err)
			return
		}
		u.Path = path.Join(u.Path, "v", v.ID)
		id := u.String()
		f.Items = append(f.Items, &feeds.Item{
			Id:          id,
			Title:       v.Title,
			Link:        &feeds.Link{Href: id},
			Description: v.Description,
			Enclosure: &feeds.Enclosure{
				Url:    id + ".mp4",
				Length: strconv.FormatInt(v.Size, 10),
				Type:   "video/mp4",
			},
			Author: &feeds.Author{
				Name:  cfg.Author.Name,
				Email: cfg.Author.Email,
			},
			Created: v.Timestamp,
		})
	}
	w.Header().Set("Cache-Control", "public, max-age=7776000")
	w.Header().Set("Content-Type", "text/xml")
	f.WriteRss(w)
}

// 404 Page handler
func (a *App) Unfoundhandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	a.Templates.ExecuteTemplate(w, "404.html", nil)
}

func (a *App) Deniedhandler(w http.ResponseWriter, r *http.Request, error *ErrorHandler) {
	// w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
	w.WriteHeader(http.StatusUnauthorized)
	a.Templates.ExecuteTemplate(w, "401.html", error)
}

// Get Logger.
func (app *App) GetLogger() *mylog.Logger {
	return app.Logger
}

// Login Handler
func (a *App) loginHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the username and password from the request body
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			a.Logger.Log.Error("Failed to parse form:", err)
			a.Deniedhandler(w, r, &ErrorHandler{Error: err.Error()})
			return
		}
		username := r.Form.Get("username")
		password := r.Form.Get("password")
		// Authenticate the user and set the session if successful
		if a.Middleware.AuthenticateAndSetSession(username, password, w, r) {
			// Authentication succeeded, redirect to the authenticated route
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		// Authentication failed, redirect to the login page
		// http.Redirect(w, r, "/login", http.StatusFound)
		a.Templates.ExecuteTemplate(w, "login.html", &ErrorHandler{Error: "Wrong username or password"})
	} else {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		a.Templates.ExecuteTemplate(w, "login.html", nil)
	}
}

// Logout Handler
func (a *App) logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := a.Sessions.Get(r, "session")
	if err != nil {
		a.FailedSession(w, r)
	}
	session.Values["authenticated"] = false
	session.Values["username"] = ""
	session.Values["SessionID"] = ""
	err = session.Save(r, w)
	if err != nil {
		a.Logger.Log.Warn("Failed to save session:", err)
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// Sign up Handler
func (a *App) signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		a.Templates.ExecuteTemplate(w, "login.html", nil)
	}
		err := r.ParseForm()
	if err != nil {
		a.Logger.Log.Error("Failed to parse sign-up form:", err)
		a.Templates.ExecuteTemplate(w, "login.html", &ErrorHandler{Error: "Failed to parse sign-up form"})
	}

	username := r.Form.Get("username")
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	a.Logger.Log.Logf("Sign-up Request from %s", r.UserAgent())
	// Validate email
	a.Mdata.Email = email
	_, err = a.Mdata.Validate()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	isUser, err := a.Middleware.SignUpAndSetSession(username, password, email, w, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		a.Templates.ExecuteTemplate(w, "login.html", &ErrorHandler{Error: err.Error()})
		return
	}
	if isUser {
		http.Redirect(w, r, "/", http.StatusFound)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		a.Templates.ExecuteTemplate(w, "login.html", &ErrorHandler{Error: err.Error()})
	}
}

// Failed Sessions
func (a *App) FailedSession(w http.ResponseWriter, r *http.Request) {
		a.Logger.Log.Warn("Failed to obtain session.")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
}

// Restricted Acesss and authentication
func (a *App) MediaAcess(w http.ResponseWriter, r *http.Request, m *media.Video) (bool, error) {
	session, err := a.Sessions.Get(r, "session")
	if err != nil {
		a.FailedSession(w, r)
		return false, err
	}
	authenticated, _ := session.Values["authenticated"].(bool)
	if m.Restricted {
		if restrictedAcc, ok := session.Values["restrictedAcc"].(bool); !ok || !restrictedAcc {
			return authenticated, errors.New("Insufficient Access")
		}
	}
	return authenticated, nil
}
