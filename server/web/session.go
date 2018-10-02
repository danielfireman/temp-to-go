package main

import (
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo"
	"github.com/labstack/echo-contrib/session"
)

const (
	loggedInSessionField = "loggedin"
	sessionName          = "session"
)

type loginHandler struct {
	userPasswd string
}

func (h *loginHandler) handle(c echo.Context) error {
	user := c.FormValue("user")
	passwd := c.FormValue("password")
	if user+passwd == h.userPasswd {
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
	return c.Render(http.StatusOK, "login", struct {
		LoginError string
	}{
		LoginError: "Invalid credentials. Try again.",
	})
}

type logoutHandler struct {
}

func (h *logoutHandler) handle(c echo.Context) error {
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

func loginCheckMiddleware(in echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		sess, err := session.Get(sessionName, c)
		if err != nil {
			c.Logger().Errorf("Err getting session (%s): %q\n", sessionName, err)
			return c.NoContent(http.StatusInternalServerError)
		}
		_, isLoggedIn := sess.Values[loggedInSessionField]
		if err != nil {
			c.Logger().Errorf("Err checking login: %q\n", err)
			return c.NoContent(http.StatusInternalServerError)
		}
		if isLoggedIn {
			return in(c)
		}
		errorResponse := struct {
			Error string `json:"error"`
		}{
			Error: "Access restricted, please log in first",
		}
		if c.Request().Header.Get("Content-type") == "application/json" {
			return c.JSON(http.StatusForbidden, errorResponse)
		}
		return c.Render(http.StatusOK, "error", errorResponse)
	}
}
