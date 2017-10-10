package app

import (
	"fmt"
	"time"

	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	"github.com/kataras/iris/sessions"
	"github.com/speedwheel/dropzone/app/controllers"
)

// Application is our application wrapper and bootstrapper, keeps our settings.
type Application struct {
	*iris.Application

	Name      string
	Owner     string
	SpawnDate time.Time

	Sessions *sessions.Sessions
}

// NewApplication returns a new named Application.
func NewApplication(name, owner string) *Application {
	return &Application{
		Name:        name,
		Owner:       owner,
		Application: iris.New(),
		SpawnDate:   time.Now(),
	}
}

// begin sends the app's identification info.
func (app *Application) begin(ctx iris.Context) {
	// response headers
	ctx.Header("App-Name", app.Name)
	ctx.Header("App-Owner", app.Owner)
	ctx.Header("App-Since", time.Since(app.SpawnDate).String())

	ctx.Header("Server", "Iris: https://iris-go.com")

	// view data if ctx.View or c.Tmpl = "$page.html" will be called next.
	ctx.ViewData("AppName", app.Name)
	ctx.ViewData("AppOwner", app.Owner)
	ctx.Next()
}

// SetupViews loads the templates.
func (app *Application) SetupViews(viewsDir string) {
	app.RegisterView(iris.HTML(viewsDir, ".html").Reload(true).Layout("shared/layout.html"))
}

// SetupErrorHandlers prepares the http error handlers (>=400).
// Remember that error handlers in Iris have their own middleware ecosystem
// so the route's middlewares are not running when an http error happened.
// So if we want a logger we have to re-create one, here we will customize that logger as well.
func (app *Application) SetupErrorHandlers() {
	httpErrStatusLogger := logger.New(logger.Config{
		Status:            true,
		IP:                true,
		Method:            true,
		Path:              true,
		MessageContextKey: "message",
		LogFunc: func(now time.Time, latency time.Duration,
			status, ip, method, path string,
			message interface{}) {

			line := fmt.Sprintf("%v %4v %s %s %s", status, latency, ip, method, path)

			if message != nil {
				line += fmt.Sprintf(" %v", message)
			}
			app.Logger().Warn(line)
		},
	})

	app.OnAnyErrorCode(app.begin, httpErrStatusLogger, func(ctx iris.Context) {
		err := iris.Map{
			"app":     app.Name,
			"status":  ctx.GetStatusCode(),
			"message": ctx.Values().GetString("message"),
		}

		if jsonOutput, _ := ctx.URLParamBool("json"); jsonOutput {
			ctx.JSON(err)
			return
		}

		ctx.ViewData("Err", err)
		ctx.ViewData("Title", "Error")
		ctx.View("shared/error.html")
	})
}

// SetupRouter registers the available routes from the "controllers" package.
func (app *Application) SetupRouter() {
	app.Use(recover.New())
	app.Use(app.begin)
	app.Use(iris.Gzip)

	app.Favicon("./public/favicon.ico")
	app.StaticWeb("/public", "./public")

	app.Use(logger.New())

	app.Controller("/", new(controllers.IndexController))
}

// Instance is our global application bootstrap instance.
var Instance = NewApplication("Iris + Dropzone", "info@domain.com")

// Boot starts our default instance appolication.
func Boot(runner iris.Runner, configurators ...iris.Configurator) {
	Instance.SetupViews("./app/views")

	Instance.SetupErrorHandlers()
	Instance.SetupRouter()

	Instance.Run(runner, configurators...)
}
