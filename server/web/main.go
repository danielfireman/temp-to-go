package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/danielfireman/temp-to-go/server/tsmongo"
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
	tsmongoSession, err := tsmongo.Dial(mgoURI)
	if err != nil {
		log.Fatalf("Error connecting to StatusDB: %s", mgoURI)
	}
	defer tsmongoSession.Close()
	fanService := tsmongo.NewFanService(tsmongoSession)
	bedroomService := tsmongo.NewBedroomService(tsmongoSession)
	weatherService := tsmongo.NewWeatherService(tsmongoSession)

	publicHTML := filepath.Join(os.Getenv("PUBLIC_HTML"))

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

	// Public Routes.
	bedroomAPIHandler := bedroomAPIHandler{key, bedroomService}
	loginHandler := loginHandler{userPasswd}
	e.GET("/", func(c echo.Context) error {
		return c.Render(http.StatusOK, "login", struct {
			LoginError string
		}{
			LoginError: "",
		})
	})

	e.File("/favicon.ico", filepath.Join(publicHTML, "favicon.ico"))
	e.POST("/indoortemp", bedroomAPIHandler.handlePost)
	e.POST("/login", loginHandler.handle)

	// Routes which should only be accessed after login.
	restricted := e.Group(restrictedPath, loginCheckMiddleware)

	restrictedMainHandler := restrictedMainHandler{fanService}
	weatherHandler := weatherHandler{weatherService}
	fanHandler := fanHandler{fanService}
	logoutHandler := logoutHandler{}
	restricted.GET("", restrictedMainHandler.handle)
	restricted.POST("/fan", fanHandler.handle)
	restricted.POST("/logout", logoutHandler.handle)
	restricted.GET("/weather", weatherHandler.handle)
	restricted.GET("/indoortemp", bedroomAPIHandler.handleGet)

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
