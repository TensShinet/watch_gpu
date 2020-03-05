package routers

import (
	"github.com/TensShinet/watch_gpu/server/controllers"
	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/gpu_information", &controllers.GpuController{})
}
