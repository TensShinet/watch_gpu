package controllers

import (
	"io/ioutil"

	"github.com/astaxie/beego"
)

type MainController struct {
	beego.Controller
}

func (this *MainController) Get() {
	text, err := ioutil.ReadFile("./views/index.html")
	if err != nil {
		this.Data["json"] = valueTypeError{
			"valueError",
			400,
		}
		this.ServeJSON()
		return
	}

	this.Ctx.WriteString(string(text))
}
