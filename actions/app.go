package actions

import (
	"sync"
	"time"

	"coke/models"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo-pop/v3/pop/popmw"
	"github.com/gobuffalo/envy"
	forcessl "github.com/gobuffalo/mw-forcessl"
	paramlogger "github.com/gobuffalo/mw-paramlogger"
	tokenauth "github.com/gobuffalo/mw-tokenauth"
	"github.com/gobuffalo/x/sessions"
	"github.com/golang-jwt/jwt/v4"
	"github.com/rs/cors"
	"github.com/unrolled/secure"
)

type Response struct {
	Data   interface{} `json:"data"`
	Errors interface{} `json:"errors"`
	Status string      `json:"status"`
}

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")

var (
	app     *buffalo.App
	appOnce sync.Once
)

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
	appOnce.Do(func() {
		app = buffalo.New(buffalo.Options{
			Env:          ENV,
			SessionStore: sessions.Null{},
			PreWares: []buffalo.PreWare{
				cors.Default().Handler,
			},
			SessionName: "_coke_session",
		})

		// Automatically redirect to SSL
		app.Use(forceSSL())

		// Log request parameters (filters apply).
		app.Use(paramlogger.ParameterLogger)

		// Set the request content type to JSON
		// app.Use(contenttype.Set("application/json"))

		// Wraps each request in a transaction.
		//   c.Value("tx").(*pop.Connection)
		// Remove to disable this.
		app.Use(popmw.Transaction(models.DB))
		app.GET("/", HomeHandler)

		app.Use(AuthJwt())
		app.Use(SetCurrentUser)
		app.Middleware.Skip(AuthJwt(), AuthCreate)

		ur := UserResource{}
		app.GET("/users", ur.Index)
		app.GET("/users/{user_id}", ur.Show)
		app.POST("/users", ur.Store)
		app.POST("/auth", AuthCreate)
		app.GET("/auth", AuthIndex)

	})

	return app
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

func AuthJwt() buffalo.MiddlewareFunc {
	return tokenauth.New(tokenauth.Options{
		SignMethod: jwt.SigningMethodHS256,
	})
}

func SetCurrentUser(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		user := &models.User{}
		cv := c.Value("claims")
		if cv == nil {
			return next(c)
		}

		claims := cv.(jwt.MapClaims)
		userId := claims["user_id"].(float64)
		exp := int64(claims["exp"].(float64))

		if time.Now().Unix() > exp {
			return c.Render(401, r.JSON(Response{
				Errors: "Token is expired",
			}))
		}

		err := models.DB.Find(user, userId)
		if err == nil {
			c.Set("auth", user)
		}

		return next(c)
	}
}
