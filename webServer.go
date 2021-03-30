package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WebServer struct {
	router *gin.Engine
	node   *Node
}

func initServer(n *Node) *WebServer {
	return &WebServer{
		router: gin.Default(),
		node:   n,
	}
}
func (w WebServer) initializeRoutes() {
	// Handle the index route
	w.router.GET("/", func(c *gin.Context) {

		// Call the HTML method of the Context to render a template
		c.HTML(
			// Set the HTTP status to 200 (OK)
			http.StatusOK,
			// Use the index.html template
			"index.html",
			// Pass the data that the page uses (in this case, 'title')
			gin.H{
				"title": "Home Page",
			},
		)
	})
	api := w.router.Group("/api")
	{
		api.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": w.node.NodeCore.String(),
			})
		})
		//TODO: functions of api
		// readkey: get k-value of node's data
		// readall: get all k-values of nodes
		// insert: lookup where to put key and send store to nodes for key
		// dont forget to cache titles of keys into personal
		// search: lookup key and return value
		// v1.GET("read/:key", GetKeyinData)
		// v1.GET("readAll", GetData)
		// v1.POST("insert/:key", InsertKey)
		// v1.GET("search/:key", SearchKey)
	}

}

func (w WebServer) runWebServer(port int) {

	w.router.LoadHTMLGlob("templates/*")
	w.initializeRoutes()
	w.router.Run(fmt.Sprintf(":%d", port))
}

// func GetKeyinData(c *gin.Context) {
// 	key := c.Params.ByName("key")

// 	if n.Data[key] {
// 		c.JSON(http.StatusOK, n.Data[key])
// 	}

// 	if err != nil {
// 		c.AbortWithStatus(http.StatusNotFound)
// 	} else {
// 		c.JSON(http.StatusOK, todo)
// 	}
// }
