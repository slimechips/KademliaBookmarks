package main

import (
	"fmt"
	"log"
	"net/http"

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
				"hi": w.node.NodeCore.String(),
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
					"Read Key": "key not found:" + str})
			}
		})
		// readkey: get k-value of node's data
		api.GET("/readAll", func(c *gin.Context) {
			log.Print("Trying to read all my data\n")
			s := ""
			for k, v := range w.node.Data {
				s += fmt.Sprintf("%s:%s\n", k.String(), v)
			}
			c.JSON(http.StatusOK, gin.H{
				"Read All Keys": s})
		})
		// search: lookup key and return value (need to implement return of string in FindValueByKey)
		api.POST("/searchValueByKey", func(c *gin.Context) {
			str := c.PostForm("searchkey")
			key := ConvertStringToID(str)
			log.Printf("searchValueByKey: %s\n", str)
			s := w.node.FindValueByKey(key)
			c.JSON(http.StatusOK, gin.H{
				"Search Value By Key": s})
		})

		// insert: lookup where to put key and send store to nodes for key
		api.POST("/insert", func(c *gin.Context) {
			key := c.PostForm("insertkey")
			val := c.PostForm("insertval")
			keyID := ConvertStringToID(key)
			log.Printf("key:%s, val:%s\n", keyID, val)
			nodeCores := w.node.KNodesLookUp(keyID)
			nodestr := ""
			for _, nc := range nodeCores {
				nodestr += fmt.Sprintf("%s:%s\t", nc.GUID.String(), nc.IP.String())
			}
			log.Printf("Storing at nodecores: %s\n", nodestr)
			w.node.StoreInNodes(nodeCores, keyID, val)
			c.JSON(http.StatusOK, gin.H{
				"inserted value": fmt.Sprintf("%s at %s", val, nodestr)})
		})
		// search: lookup key and return value (need to implement return of string in FindValueByKey)
		api.GET("/readNeighbors", func(c *gin.Context) {
			s := ""
			for i := 0; i < ID_LENGTH*8; i++ {
				bucket := w.node.RoutingTable.Buckets[i]
				for e := bucket.Front(); e != nil; e = e.Next() {
					s += e.Value.(*NodeCore).String()
				}
			}
			log.Printf("readNeighbors: %s\n", s)
			c.JSON(http.StatusOK, gin.H{
				"Search Value By Key": s})
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
