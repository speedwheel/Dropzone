package controllers

import (
	"github.com/kataras/iris"
	"github.com/kataras/iris/mvc"
	"github.com/speedwheel/dropzone/app/domain"
)

type IndexController struct {
	mvc.C
}

var indexView = mvc.View{
	Name: "index.html",
	Data: map[string]interface{}{
		"Title":     "Upload Images Page",
		"MyMessage": "Welcome to Iris + Dropzone",
	},
}

func (c *IndexController) Get() mvc.Result {
	return indexView
}

func (c *IndexController) PostUpload() {
	file, info, err := c.Ctx.FormFile("file")
	if err != nil {
		c.Ctx.StatusCode(iris.StatusInternalServerError)
		c.Ctx.Application().Logger().Warnf("Error while uploading: %v", err.Error())
		return
	}

	defer file.Close()

	domain.Upload(file, info, c.Ctx)
}

func (c *IndexController) DeleteRemoveBy(name string) bool {
	if name == "" {
		return false
	}
	return domain.Remove(name)
}

func (c *IndexController) GetUploads() []domain.UploadedFile {
	return domain.RefreshUploadFolder().Items
}
