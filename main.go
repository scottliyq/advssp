package main

import (
	_ "advapi/docs"
	_ "advapi/routers"
	"advapi/tools"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	tools.Init("ip.dat")
	tools.InitRequestPool("127.0.0.1:11300", "MDADV_REQUEST_DEVICE_LOG")
	orm.RegisterDriver("mysql", orm.DR_MySQL)
	orm.RegisterDataBase("default", "mysql", "root:root@tcp(127.0.0.1:8889)/addata?charset=utf8")
}

func main() {
	if beego.RunMode == "dev" {
		beego.DirectoryIndex = true
		beego.StaticDir["/swagger"] = "swagger"
	}
	beego.ViewsPath = "view"
	beego.Run()
}
