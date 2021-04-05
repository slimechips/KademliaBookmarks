package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type WebServer struct {
	router *gin.Engine
	node   *Node
}

func initServer(n *Node) *WebServer {
	r := gin.Default()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	return &WebServer{
		router: r,
		node:   n,
	}
}
func (w WebServer) initializeRoutes() {
	// Handle the index route
	w.router.GET("/", func(c *gin.Context) {
		dataLinks := w.node.getData()
		// Call the HTML method of the Context to render a template
		c.HTML(
			// Set the HTTP status to 200 (OK)
			http.StatusOK,
			// Use the index.html template
			"index.html",
			// Pass the data that the page uses (in this case, 'title')
			gin.H{
				"title":   "KademliaBM",
				"payload": dataLinks,
			},
		)
	})
	api := w.router.Group("/api")
	{
		//home page
		api.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": w.node.NodeCore.String(),
			})
		})
		// readkey: get k-value of node's data
		api.POST("/readKey", func(c *gin.Context) {
			str := c.PostForm("readkey")
			key := ConvertStringToID(str)
			log.Printf("readKeyyy: %s\n", str)
			if val, ok := w.node.Data[key]; ok {
				c.JSON(http.StatusOK, gin.H{
					"Read Key": val})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"Read Key": "notfound:" + str})
			}
		})
		// readkey: get k-value of node's data
		api.POST("/readAllKey", func(c *gin.Context) {
			s := ""
			for k, v := range w.node.Data {
				s += fmt.Sprintf("%s:%s", k.String(), v)
			}
			c.JSON(http.StatusOK, gin.H{
				"Read All Keys": s})
		})
		// search: lookup key and return value (need to implement return of string in FindValueByKey)
		api.POST("/searchValueByKey", func(c *gin.Context) {
			str := c.PostForm("searchkey")
			key := ConvertStringToID(str)
			log.Printf("searchValueByKey: %s\n", str)
			w.node.FindValueByKey(key)
			c.JSON(http.StatusOK, gin.H{
				"Search Value By Key": str})
		})

		// insert: lookup where to put key and send store to nodes for key
		api.POST("/insert", func(c *gin.Context) {
			str := c.PostForm("insertval")
			key := ConvertStringToID(str)
			nodeCores := w.node.KNodesLookUp(key)
			nodestr := ""
			for _, nc := range nodeCores {
				nodestr += fmt.Sprintf("%s:%s\t", nc.GUID.String(), nc.IP.String())
			}
			log.Printf("Storing at nodecores: %s\n", nodestr)
			w.node.StoreInNodes(nodeCores, key, str)
			<-time.NewTimer(time.Duration(10) * time.Second).C
			c.JSON(http.StatusOK, gin.H{
				"inserted value": str})
		})
		//TODO: functions of api

		// dont forget to cache titles of keys into personal

	}

}

func (w WebServer) runWebServer(port int) {

	w.router.LoadHTMLGlob("templates/*")
	w.initializeRoutes()
	w.router.Run(fmt.Sprintf(":%d", port))
}
