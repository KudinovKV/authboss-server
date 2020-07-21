package web

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/KudinovKV/authboss-server/database"
	"github.com/aarondl/tpl"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-pg/pg"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/justinas/nosurf"
	zl "github.com/rs/zerolog/log"
	abclientstate "github.com/volatiletech/authboss-clientstate"
	abrenderer "github.com/volatiletech/authboss-renderer"
	"github.com/volatiletech/authboss/v3"
	_ "github.com/volatiletech/authboss/v3/auth"
	"github.com/volatiletech/authboss/v3/confirm"
	"github.com/volatiletech/authboss/v3/defaults"
	"github.com/volatiletech/authboss/v3/lock"
	_ "github.com/volatiletech/authboss/v3/logout"
	_ "github.com/volatiletech/authboss/v3/recover"
	_ "github.com/volatiletech/authboss/v3/register"
	"github.com/volatiletech/authboss/v3/remember"
)

type Server struct {
	ab        *authboss.Authboss
	r         *chi.Mux
	templates tpl.Templates
}

// newServer create new copy of server struct
func newServer() *Server {
	return &Server{
		ab: authboss.New(),
		r:  chi.NewRouter(),
		templates: tpl.Must(tpl.Load("web", "", "layout.html.tpl", template.FuncMap{
			"yield": func() string { return "" },
		})),
	}
}

// initCookie generate cookie password and set cookie options
func initCookie() abclientstate.CookieStorer {
	cookieStoreKey := securecookie.GenerateRandomKey(64)
	cookieStore := abclientstate.NewCookieStorer(cookieStoreKey, nil)
	cookieStore.HTTPOnly = true
	cookieStore.Secure = true
	cookieStore.MaxAge = int(24 * time.Hour / time.Second)
	return cookieStore
}

// initSession generate session password and set session options
func initSession() abclientstate.SessionStorer {
	sessionStoreKey := securecookie.GenerateRandomKey(64)
	sessionStore := abclientstate.NewSessionStorer("AB_Server_Session", sessionStoreKey, nil)
	cstore := sessionStore.Store.(*sessions.CookieStore)
	cstore.Options.HttpOnly = false
	cstore.Options.Secure = false
	cstore.MaxAge(int((30 * 24 * time.Hour) / time.Second))
	return sessionStore
}

// initAuthBoss init copy of authboss
func (s *Server) initAuthBoss(listen string, db *pg.DB) {
	s.ab.Config.Paths.RootURL = listen
	s.ab.Config.Modules.LogoutMethod = "GET"
	s.ab.Config.Storage.Server = database.NewStorer(db)
	s.ab.Config.Storage.SessionState = initSession()
	s.ab.Config.Storage.CookieState = initCookie()
	s.ab.Config.Core.ViewRenderer = abrenderer.NewHTML("/auth", "ab_views")
	s.ab.Config.Core.MailRenderer = abrenderer.NewEmail("/auth", "ab_views")
	s.ab.Config.Modules.RegisterPreserveFields = []string{"email", "name", "role"}
	s.ab.Config.Modules.ResponseOnUnauthed = authboss.RespondRedirect
	s.ab.Config.Modules.TwoFactorEmailAuthRequired = false
	defaults.SetCore(&s.ab.Config, false, false)
	emailRule := defaults.Rules{
		FieldName: "email", Required: true,
		MatchError: "Must be a valid e-mail address",
		MustMatch:  regexp.MustCompile(`.*@.*\.[a-z]+`),
	}
	passwordRule := defaults.Rules{
		FieldName: "password", Required: true,
		MinLength: 4,
	}
	nameRule := defaults.Rules{
		FieldName: "name", Required: true,
		MinLength: 2,
	}
	roleRule := defaults.Rules{
		FieldName: "role",
		Required:  true,
	}
	s.ab.Config.Core.BodyReader = defaults.HTTPBodyReader{
		ReadJSON: false,
		Rulesets: map[string][]defaults.Rules{
			"register":    {emailRule, passwordRule, nameRule, roleRule},
			"recover_end": {passwordRule},
		},
		Confirms: map[string][]string{
			"register":    {"password", authboss.ConfirmPrefix + "password"},
			"recover_end": {"password", authboss.ConfirmPrefix + "password"},
		},
		Whitelist: map[string][]string{
			"register": {"email", "name", "password", "role"},
		},
	}
	if err := s.ab.Init(); err != nil {
		zl.Error().Err(err).Msg("Can't init authboss")
	}
}

// InitServer init server and set server options
func InitServer(listen string, db *pg.DB) {
	s := newServer()
	s.initAuthBoss(listen, db)
	s.r.Use(middleware.Timeout(60 * time.Second))
	s.r.Use(middleware.Logger, nosurfing, s.ab.LoadClientStateMiddleware, remember.Middleware(s.ab), s.dataInjector)
	s.r.Group(func(mux chi.Router) {
		mux.Use(authboss.Middleware2(s.ab, authboss.RequireNone, authboss.RespondUnauthorized), lock.Middleware(s.ab), confirm.Middleware(s.ab))
		mux.MethodFunc("GET", "/foo", s.foo)
		mux.MethodFunc("GET", "/bar", s.bar)
		mux.MethodFunc("GET", "/sigma", s.sigma)
	})
	s.r.Group(func(mux chi.Router) {
		mux.Use(authboss.ModuleListMiddleware(s.ab))
		mux.Mount("/auth", http.StripPrefix("/auth", s.ab.Config.Core.Router))
	})
	s.r.Get("/", s.index)
	s.r.Get("/healthcheck", s.healthcheck)
	optionsHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-CSRF-TOKEN", nosurf.Token(r))
		w.WriteHeader(http.StatusOK)
	}
	routes := []string{"login", "logout", "recover", "recover/end", "register"}
	s.r.MethodFunc("OPTIONS", "/*", optionsHandler)
	for _, r := range routes {
		s.r.MethodFunc("OPTIONS", "/auth/"+r, optionsHandler)
	}
	zl.Debug().Msgf("Start server at %v", listen)
	err := http.ListenAndServe(listen, s.r)
	if err != nil {
		zl.Error().Err(err).Msg("Can't start server")
	}
}

// dataInjector handle and write context in current ResponseWriter
func (s Server) dataInjector(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := s.layoutData(w, &r)
		zl.Debug().Msgf("%v", data)
		r = r.WithContext(context.WithValue(r.Context(), authboss.CTXKeyData, data))
		handler.ServeHTTP(w, r)
	})
}

// layoutData get user auth settings
func (s Server) layoutData(w http.ResponseWriter, r **http.Request) authboss.HTMLData {
	currentUserName := ""
	userInter, err := s.ab.LoadCurrentUser(r)
	if userInter != nil && err == nil {
		currentUserName = userInter.(*database.User).Name
	}

	return authboss.HTMLData{
		"loggedin":          userInter != nil,
		"current_user_name": currentUserName,
		"csrf_token":        nosurf.Token(*r),
		"flash_success":     authboss.FlashSuccess(w, *r),
		"flash_error":       authboss.FlashError(w, *r),
	}
}

// nosurfing generate and check correct csrf token
func nosurfing(h http.Handler) http.Handler {
	surfing := nosurf.New(h)
	surfing.SetFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		zl.Debug().Msgf("Failed to validate CSRF token:", nosurf.Reason(r))
		w.WriteHeader(http.StatusBadRequest)
	}))
	return surfing
}

// healthcheck check is server alive
func (s Server) healthcheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("Server is working"))
	if err != nil {
		zl.Fatal().Err(err).Msg("Server is down")
	}
}

// index handle main page
func (s Server) index(w http.ResponseWriter, r *http.Request) {
	s.render(w, r, "index", authboss.HTMLData{})
}

// render gets and puts data to template and render it
func (s Server) render(w http.ResponseWriter, r *http.Request, name string, data authboss.HTMLData) {
	var current authboss.HTMLData
	dataIntf := r.Context().Value(authboss.CTXKeyData)
	if dataIntf == nil {
		current = authboss.HTMLData{}
	} else {
		current = dataIntf.(authboss.HTMLData)
	}
	current.MergeKV("csrf_token", nosurf.Token(r))
	current.Merge(data)
	err := s.templates.Render(w, name, current)
	if err == nil {
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = fmt.Fprintln(w, "Error occurred rendering template:", err)
}

// validation checks if the current user role belongs to a whitelisted role
func (s Server) validation(w http.ResponseWriter, r *http.Request, correctRole []string) {
	abUser, err := s.ab.LoadCurrentUser(&r)
	if err != nil {
		zl.Warn().Err(err).Msgf("Can't load user")
		ro := authboss.RedirectOptions{
			Code:         http.StatusTemporaryRedirect,
			RedirectPath: "/",
			Failure:      "Can't load your user.",
		}
		err := s.ab.Core.Redirector.Redirect(w, r, ro)
		if err != nil {
			zl.Warn().Err(err).Msgf("Can't redirect to %v", s.ab.Config.Paths.NotAuthorized)
		}
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	}
	currentUserRole := abUser.(*database.User).GetRole()
	for _, currentRole := range correctRole {
		if strings.EqualFold(currentUserRole, currentRole) {
			return
		}
	}
	zl.Warn().Err(err).Msgf("User %v have incorrect role %v", abUser.(*database.User).GetEmail(), currentUserRole)
	ro := authboss.RedirectOptions{
		Code:         http.StatusTemporaryRedirect,
		RedirectPath: "/",
		Failure:      `Your role don't support this resource. Only "Administrator" can load this page.`,
	}
	err = s.ab.Core.Redirector.Redirect(w, r, ro)
	if err != nil {
		zl.Warn().Err(err).Msgf("Can't redirect to %v", s.ab.Config.Paths.NotAuthorized)
	}
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// foo handler
func (s Server) foo(w http.ResponseWriter, r *http.Request) {
	s.validation(w, r, []string{"User", "Administrator"})
	s.render(w, r, "resource", authboss.HTMLData{"resource": "Foo"})
}

// bar handler
func (s Server) bar(w http.ResponseWriter, r *http.Request) {
	s.validation(w, r, []string{"User", "Administrator"})
	s.render(w, r, "resource", authboss.HTMLData{"resource": "Bar"})
}

// sigma handler
func (s Server) sigma(w http.ResponseWriter, r *http.Request) {
	s.validation(w, r, []string{"Administrator"})
	s.render(w, r, "resource", authboss.HTMLData{"resource": "Sigma"})
}
