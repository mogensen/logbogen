package actions

import (
	_ "logbogen/achievementevents"
	"logbogen/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/events"
	forcessl "github.com/gobuffalo/mw-forcessl"
	paramlogger "github.com/gobuffalo/mw-paramlogger"
	"github.com/gobuffalo/plush/v4"
	"github.com/unrolled/secure"

	"github.com/gobuffalo/buffalo-pop/v2/pop/popmw"
	csrf "github.com/gobuffalo/mw-csrf"
	i18n "github.com/gobuffalo/mw-i18n"
	"github.com/gobuffalo/packr/v2"

	"github.com/markbates/goth/gothic"
)

var app *buffalo.App

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")

// T is used to handle all translations
var T *i18n.Translator

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
//
// Routing, middleware, groups, etc... are declared TOP -> DOWN.
// This means if you add a middleware to `app` *after* declaring a
// group, that group will NOT have that new middleware. The same
// is true of resource declarations as well.
//
// It also means that routes are checked in the order they are declared.
// `ServeFiles` is a CATCH-ALL route, so it should always be
// placed last in the route declarations, as it will prevent routes
// declared after it to never be called.
func App() *buffalo.App {
	if app == nil {
		app = buffalo.New(buffalo.Options{
			Env:         ENV,
			SessionName: "_logbogen_session",
		})

		// Automatically redirect to SSL
		app.Use(forceSSL())

		// Log request parameters (filters apply).
		app.Use(paramlogger.ParameterLogger)

		// Protect against CSRF attacks. https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)
		// Remove to disable this.
		app.Use(csrf.New)

		// Wraps each request in a transaction.
		//  c.Value("tx").(*pop.Connection)
		// Remove to disable this.
		app.Use(popmw.Transaction(models.DB))

		// Setup and use translations:
		app.Use(translations())

		app.Use(plushTimeFormat())

		app.GET("/", HomeHandler)

		auth := app.Group("/auth")
		app.Use(SetCurrentUser)
		app.Use(Authorize)

		app.Middleware.Skip(Authorize, HomeHandler, UsersNew, UsersCreate, AuthCreate, SwitchLanguage)
		app.Resource("/users_images", UsersImagesResource{})

		app.POST("/language", SwitchLanguage)

		app.GET("/users/new", UsersNew)
		app.POST("/users", UsersCreate)
		app.POST("/signin", AuthCreate)

		app.GET("/pendingactivities", PendingActivitiesList)
		app.GET("/climbingactivities/{climbingactivity_id}/clone", CloneClimbingActivity)

		app.GET("/users/{user_id}", UserShow)
		app.GET("/users/{user_id}", UserEdit)

		bah := buffalo.WrapHandlerFunc(gothic.BeginAuthHandler)
		auth.GET("/{provider}", bah)
		auth.DELETE("", AuthDestroy)
		auth.Middleware.Skip(Authorize, bah, AuthCallback)
		auth.GET("/{provider}/callback", AuthCallback)
		app.Resource("/climbingactivities", ClimbingactivitiesResource{})

		app.ServeFiles("/", assetsBox) // serve files from the public directory
	}

	e := events.Event{
		Kind: "logbogen:achievements:updateall",
	}
	events.Emit(e)

	return app
}

// plushTimeFormat sets default time format for all plush templates
func plushTimeFormat() buffalo.MiddlewareFunc {
	plush.DefaultTimeFormat = "02 Jan 2006"
	return T.Middleware()
}

// translations will load locale files, set up the translator `actions.T`,
// and will return a middleware to use to load the correct locale for each
// request.
// for more information: https://gobuffalo.io/en/docs/localization
func translations() buffalo.MiddlewareFunc {
	var err error
	if T, err = i18n.New(packr.New("app:locales", "../locales"), "en-US"); err != nil {
		app.Stop(err)
	}
	return T.Middleware()
}

// forceSSL will return a middleware that will redirect an incoming request
// if it is not HTTPS. "http://example.com" => "https://example.com".
// This middleware does **not** enable SSL. for your application. To do that
// we recommend using a proxy: https://gobuffalo.io/en/docs/proxy
// for more information: https://github.com/unrolled/secure/
func forceSSL() buffalo.MiddlewareFunc {
	return forcessl.Middleware(secure.Options{
		SSLRedirect:     ENV == "production",
		SSLProxyHeaders: map[string]string{"X-Forwarded-Proto": "https"},
	})
}
