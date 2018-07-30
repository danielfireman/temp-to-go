package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/danielfireman/temp-to-go/server/status"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"

	htmlTemplate "html/template"
)

const (
	restrictedPath = "/restricted"
)

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

	publicHTML := filepath.Join(os.Getenv("PUBLIC_HTML"))
	fan := sdb.Fan()

	// Initializing web framework.
	e := echo.New()

	// Registering templates.
	// https://echo.labstack.com/guide/templates
	e.Renderer = &template{
		templates: htmlTemplate.Must(htmlTemplate.ParseGlob(filepath.Join(publicHTML, "templates", "*.html"))),
	}

	// Middlewares.
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(session.Middleware(sessions.NewCookieStore(key)))
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{Level: 5}))
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"https://mybedroom.live"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		Skipper: func(c echo.Context) bool {
			return strings.HasPrefix(c.Request().Host, "localhost")
		},
	}))

	// Public Routes.
	bedroomAPIHandler := bedroomAPIHandler{key, sdb}
	loginHandler := loginHandler{userPasswd}
	e.File("/", filepath.Join(publicHTML, "index.html"))
	e.File("/favicon.ico", filepath.Join(publicHTML, "favicon.ico"))
	e.Static("/", publicHTML)
	e.POST("/indoortemp", bedroomAPIHandler.handle)
	e.POST("/login", loginHandler.handle)

	// Routes which should only be accessed after login.
	restricted := e.Group(restrictedPath, loginCheckMiddleware)

	restrictedMainHandler := restrictedMainHandler{fan}
	weatherHandler := weatherHandler{sdb}
	fanHandler := fanHandler{fan}
	logoutHandler := logoutHandler{}
	restricted.GET("", restrictedMainHandler.handle)
	restricted.POST("/fan", fanHandler.handle)
	restricted.POST("/logout", logoutHandler.handle)
	restricted.GET("/weather", weatherHandler.handle)

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

type template struct {
	templates *htmlTemplate.Template
}

func (t *template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
