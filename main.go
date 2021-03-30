package main

import (
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"

	N "./node"
)

const subnetStart = "172.16.238."

var router *gin.Engine

func main() {

	file, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	log.SetOutput(file)
	log.Println("Log file started")
	node0 := N.NewNode(true)
	go node0.Start(subnetStart + "1")
	<-time.NewTimer(time.Duration(10) * time.Second).C

	if node0.NodeCore.IP.String() == "172.16.238.1" {
		log.Println("We're node 1. Starting test case by storing")
		<-time.NewTimer(time.Duration(10) * time.Second).C
		key := N.ConvertStringToID("123")
		nodeCores := node0.KNodesLookUp(key)
		log.Printf("Node Cores: %d\n", len(nodeCores))
		node0.StoreInNodes(nodeCores, key, "Prof Sudipta rocks")
		<-time.NewTimer(time.Duration(10) * time.Second).C
		log.Println("Finding value by key now")
		node0.FindValueByKey(key)
	}

	runtime.Goexit()
	log.Println("Exiting")

	// Set the router as the default one provided by Gin
	router = gin.Default()

	// Process the templates at the start so that they don't have to be loaded
	// from the disk again. This makes serving HTML pages very fast.
	router.LoadHTMLGlob("templates/*")

	// Define the route for the index page and display the index.html template
	// To start with, we'll use an inline route handler. Later on, we'll create
	// standalone functions that will be used as route handlers.
	router.GET("/", func(c *gin.Context) {

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

	// Start serving the application
	router.Run()

}
