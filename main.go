package main

import (
	"github.com/kataras/iris"

	"github.com/speedwheel/dropzone/app"
)

func main() {
	app.Boot(iris.Addr(":8080"), iris.WithoutServerError(iris.ErrServerClosed), iris.WithoutVersionChecker)
}
