package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/danielfireman/temp-to-go/server/status"
	"github.com/danielfireman/temp-to-go/server/web/bedroomapi"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"
)

const restrictedPath = "/restricted"

func main() {
	userPasswd := os.Getenv("USERPASSWD")
	if userPasswd == "" {
		log.Fatal("USERPASSWD must be set.")
	}
	key := []byte(os.Getenv("ENCRYPTION_KEY"))
	if len(key) != 32 {
		log.Fatalf("ENCRYPTION_KEY must be 32-bytes long. Current key is \"%s\" which is %d bytes long.", key, len(key))
	}
	mgoURI := os.Getenv("MONGODB_URI")
	if mgoURI == "" {
		log.Fatalf("Invalid MONGODB_URI: %s", mgoURI)
	}
	sdb, err := status.DialDB(mgoURI)
	if err != nil {
		log.Fatalf("Error connecting to StatusDB: %s", mgoURI)
	}
	defer sdb.Close()

	// Middlewares.
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(session.Middleware(sessions.NewCookieStore(key)))

	// Public Routes.
	publicHTML := filepath.Join(os.Getenv("PUBLIC_HTML"))
	e.File("/", filepath.Join(publicHTML, "index.html"))
	e.Static("/public", publicHTML)
	e.POST("/indoortemp", bedroomapi.TempHandlerFunc(key, sdb))
	e.POST("/login", loginHandlerFunc(userPasswd))

	// Routes which should only be accessed after login.
	restricted := e.Group(restrictedPath, restrictedMiddleware)
	restricted.GET("", mainPage)
	restricted.POST("/logout", logoutHandler)

	// Starting server.
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalf("Invalid PORT: %s", port)
	}
	s := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
	}
	e.Logger.Fatal(e.StartServer(s))
}

func mainPage(c echo.Context) error {
	return c.Redirect(http.StatusFound, "/public/main.html")
}

const (
	loggedInSessionField = "loggedin"
	sessionName          = "session"
)

func logoutHandler(c echo.Context) error {
	sess, err := session.Get(sessionName, c)
	if err != nil {
		c.Logger().Errorf("Err getting session: %q\n", err)
		return c.NoContent(http.StatusInternalServerError)
	}
	sess.Options = &sessions.Options{
		Path:   "/",
		MaxAge: -1, // MaxAge<0 means delete cookie immediately.
	}
	delete(sess.Values, loggedInSessionField)
	sess.Save(c.Request(), c.Response())
	return c.Redirect(http.StatusFound, "/")
}

func loginHandlerFunc(userPasswd string) echo.HandlerFunc {
	return func(c echo.Context) error {
		user := c.FormValue("user")
		passwd := c.FormValue("password")
		if user+passwd == userPasswd {
			sess, _ := session.Get(sessionName, c)
			sess.Options = &sessions.Options{
				Path:     "/",
				MaxAge:   86400 * 7,
				HttpOnly: true,
			}
			sess.Values[loggedInSessionField] = true
			sess.Save(c.Request(), c.Response())
			return c.Redirect(http.StatusFound, restrictedPath)
		}
		return c.NoContent(http.StatusForbidden)
	}
}

func isLoggedIn(c echo.Context) (bool, error) {
	sess, err := session.Get(sessionName, c)
	if err != nil {
		return false, err
	}
	_, ok := sess.Values[loggedInSessionField]
	return ok, nil
}

func restrictedMiddleware(in echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		isLoggedIn, err := isLoggedIn(c)
		if err != nil {
			c.Logger().Errorf("Err checking login: %q\n", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		if isLoggedIn {
			return in(c)
		}
		return c.NoContent(http.StatusForbidden)
	}
}
