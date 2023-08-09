package main

import (
	"github.com/gin-gonic/gin"
	"sell/controllers"
	"sell/util"
)

func main() {
	util.GetDbInstance()
	util.AutoMigrate(util.DB)
	r := gin.Default()
	r.GET("/update", controllers.Update)
	r.GET("/state", controllers.State)
	r.GET("/get_names", controllers.GetNames)
	r.Run(":8080")
}
