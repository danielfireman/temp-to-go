package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/danielfireman/temp-to-go/server/tsmongo"
	"github.com/gorilla/sessions"
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/middleware"

	htmlTemplate "html/template"
)

const (
	restrictedPath = "/restricted"
)

// Specification represents a map of enviornment variables
type Specification struct {
	UserPassword  string `envconfig:"USERPASSWD" required:"true"`
	EncryptionKey string `envconfig:"ENCRYPTION_KEY" required:"true"`

	MongodbURI string `default:"mongodb://127.0.0.1:27017/db" envconfig:"MONGODB_URI"`
	Port       string `default:"8080"`
	PublicHTML string `envconfig:"PUBLIC_HTML" default:"public"`
}

func main() {
	var spec Specification
	err := envconfig.Process("", &spec)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Printf("Environment Variable Set:\n%+v", spec)

	key := []byte(spec.EncryptionKey)
	if len(key) != 32 {
		log.Fatalf("ENCRYPTION_KEY must be 32-bytes long. Current key is \"%s\" which is %d bytes long.", key, len(key))
	}
	tsmongoSession, err := tsmongo.Dial(spec.MongodbURI)
	if err != nil {
		log.Fatalf("Error connecting to StatusDB: %s", spec.MongodbURI)
	}
	defer tsmongoSession.Close()
	fanService := tsmongo.NewFanService(tsmongoSession)
	bedroomService := tsmongo.NewBedroomService(tsmongoSession)
	weatherService := tsmongo.NewWeatherService(tsmongoSession)

	publicHTML := filepath.Join(spec.PublicHTML)

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
	loginHandler := loginHandler{spec.UserPassword}
	e.File("/", filepath.Join(publicHTML, "index.html"))
	e.File("/favicon.ico", filepath.Join(publicHTML, "favicon.ico"))
	e.Static("/", publicHTML)
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

	s := &http.Server{
		Addr:         ":" + spec.Port,
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
