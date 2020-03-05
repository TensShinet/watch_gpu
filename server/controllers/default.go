package controllers

import (
	"io/ioutil"

	"github.com/astaxie/beego"
)

type MainController struct {
	beego.Controller
}

func (main *MainController) Get() {
	text, err := ioutil.ReadFile("./views/index.html")
	if err != nil {
		main.Data["json"] = valueTypeError{
			"valueError",
			400,
		}
		main.ServeJSON()
		return
	}

	main.Ctx.WriteString(string(text))
}
