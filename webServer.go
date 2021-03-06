package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

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
		dataLinks := w.node.getCacheOnlyKeys()
		// Call the HTML method of the Context to render a template
		c.HTML(
			// Set the HTTP status to 200 (OK)
			http.StatusOK,
			// Use the index.html template
			"index.html",
			// Pass the data that the page uses (in this case, 'title')
			gin.H{
				"Title":    "BEEP_BOOP_BM",
				"NodeInfo": w.node.NodeCore.String(),
				"Payload":  dataLinks,
				"Debug":    false,
			},
		)
	})
	api := w.router.Group("/api")
	{
		//home page
		api.GET("/", func(c *gin.Context) {
			dataLinks := w.node.getCacheOnlyKeys()
			// Call the HTML method of the Context to render a template
			c.HTML(
				// Set the HTTP status to 200 (OK)
				http.StatusOK,
				// Use the index.html template
				"index.html",
				// Pass the data that the page uses (in this case, 'title')
				gin.H{
					"Title":    "BEEP_BOOP_BM",
					"NodeInfo": w.node.NodeCore.String(),
					"Payload":  dataLinks,
					"Debug":    true,
				},
			)
		})
		// readkey: get k-value of node's data
		api.POST("/readKey", func(c *gin.Context) {
			str := c.PostForm("readkey")
			if str != "" {
				str = strings.ToUpper(str)
				keyID := NewSHA1ID(str)
				log.Printf("readKey: %s -> %s -> %s \n", str, keyID, keyID.String())
				if val, ok := w.node.Data[keyID]; ok {
					log.Println("I HAVE" + val.Value)
					strs := strings.Split(val.Value, "*")
					c.JSON(http.StatusOK, gin.H{
						"ReadKey": strs})
				} else {
					c.JSON(http.StatusOK, gin.H{
						"ReadKey": "key not found:" + str})
				}
			} else {
				c.JSON(http.StatusOK, gin.H{
					"ReadFail": "Missing Key"})
			}
		})
		// readkey: get k-value of node's data
		api.GET("/searchkeybyparameter/:optional", func(c *gin.Context) {
			str := c.Param("optional")
			if str != "" {
				str = strings.ToUpper(str)
				key := NewSHA1ID(str)
				log.Printf("searchValueByKey: %s -> %s \n", str, key)
				s := w.node.FindValueByKey(key)
				if s != "value not found" {
					strs := strings.Split(s, "*")
					c.HTML(
						// Set the HTTP status to 200 (OK)
						http.StatusOK,
						// Use the folder.html template
						"search.html",
						// Pass the data that the page uses (in this case, 'title')
						gin.H{
							"Title":   str,
							"Payload": strs[1],
						},
					)
				} else {
					c.JSON(http.StatusOK, gin.H{
						"SearchFail": "Key Not Found"})
				}

			} else {
				c.JSON(http.StatusOK, gin.H{
					"SearchFail": "Missing Key"})
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

		api.POST("/searchValueByKey", func(c *gin.Context) {
			str := c.PostForm("searchkey")
			if str != "" {
				str = strings.ToUpper(str)
				key := NewSHA1ID(str)
				log.Printf("searchValueByKey: %s -> %s \n", str, key)
				s := w.node.FindValueByKey(key)
				if s != "value not found" {
					strs := strings.Split(s, "*")
					c.JSON(http.StatusOK, gin.H{
						"SearchKey": fmt.Sprintf("%s ---- %s", str, strs[1])})
				} else {
					c.JSON(http.StatusOK, gin.H{
						"SearchFail": "Key Not Found"})
				}

			} else {
				c.JSON(http.StatusOK, gin.H{
					"SearchFail": "Missing Key"})
			}

		})

		// search: lookup key and return value (need to implement return of string in FindValueByKey)
		api.POST("/search", func(c *gin.Context) {
			str := c.PostForm("searchtext")
			stype := c.PostForm("searchtype")
			log.Println("search", str, stype)
			if stype == "/folder" {
				if str != "" {
					str = strings.ToUpper("/" + str)
					folderName := str
					keyID := NewSHA1ID(str)
					log.Printf("searchFolder: %s -> %s \n", str, keyID)
					str := w.node.FindValueByKey(keyID)
					if str != "value not found" {
						folV := strings.Split(str, "!")
						keyList := make([]string, 0)
						for _, k := range folV {
							val := w.node.FindValueByKey(NewSHA1ID(k))
							keyVal := strings.Split(val, "*")
							keyList = append(keyList, keyVal[0])
						}
						c.HTML(
							// Set the HTTP status to 200 (OK)
							http.StatusOK,
							// Use the folder.html template
							"folder.html",
							// Pass the data that the page uses (in this case, 'title')
							gin.H{
								"Title":   folderName,
								"Payload": keyList,
							},
						)
					} else {
						c.JSON(http.StatusOK, gin.H{
							"SearchFolder": "folder not found"})
					}

				} else {
					c.JSON(http.StatusOK, gin.H{
						"SearchFail": "Missing Folder Field"})
				}
			} else {
				if str != "" {
					str = strings.ToUpper(str)
					key := NewSHA1ID(str)
					log.Printf("searchValueByKey: %s -> %s \n", str, key)
					s := w.node.FindValueByKey(key)
					if s != "value not found" {
						strs := strings.Split(s, "*")
						c.HTML(
							// Set the HTTP status to 200 (OK)
							http.StatusOK,
							// Use the folder.html template
							"search.html",
							// Pass the data that the page uses (in this case, 'title')
							gin.H{
								"Title":   str,
								"Payload": strs[1],
							},
						)
					} else {
						c.JSON(http.StatusOK, gin.H{
							"SearchFail": "Key Not Found"})
					}

				} else {
					c.JSON(http.StatusOK, gin.H{
						"SearchFail": "Missing Key"})
				}
			}

		})

		api.POST("/searchFolder", func(c *gin.Context) {

			folder := c.PostForm("readFol")
			if folder != "" {
				folder = strings.ToUpper("/" + folder)
				keyID := NewSHA1ID(folder)
				log.Printf("searchFolder: %s -> %s \n", folder, keyID)
				str := w.node.FindValueByKey(keyID)
				if str != "value not found" {
					folV := strings.Split(str, "!")
					keyValues := make([][]string, 0)
					for _, k := range folV {
						val := w.node.FindValueByKey(NewSHA1ID(k))
						keyValues = append(keyValues, strings.Split(val, "*"))
					}
					c.JSON(http.StatusOK, gin.H{
						"SearchFolder": keyValues})
				} else {
					c.JSON(http.StatusOK, gin.H{
						"SearchFolder": "folder not found"})
				}

			} else {
				c.JSON(http.StatusOK, gin.H{
					"SearchFail": "Missing Folder Field"})
			}

		})

		// insert: lookup where to put key and send store to nodes for key
		api.POST("/insert", func(c *gin.Context) {
			key := strings.ToUpper(c.PostForm("insertkey"))
			val := c.PostForm("insertval")
			folder := c.PostForm("insertfol")
			if key != "" && val != "" {
				//if no folder listed use EMPTY FOLDER
				if folder == "" || folder == " " {
					folder = "NOFOLDER"
				}
				folder = strings.ToUpper("/" + folder)
				keyID := NewSHA1ID(key)
				folID := NewSHA1ID(folder)
				log.Printf("insert: %s -> %s -> %s of val %s \n", key, keyID, keyID.String(), val)
				log.Printf("insert: %s in folder %s", key, folder)
				nodeCores := w.node.KNodesLookUp(keyID)
				nodestr := ""
				for _, nc := range nodeCores {
					nodestr += fmt.Sprintf("%s:%s\t", nc.GUID.String(), nc.IP.String())
				}
				log.Printf("Storing at nodecores: %s\n", nodestr)
				w.node.StoreInNodes(nodeCores, keyID, key+"*"+val)
				//find folder values
				folstr := w.node.FindValueByKey(folID)
				if folstr != "value not found" {
					folV := strings.Split(folstr, "!")
					ap := true
					for _, Value := range folV {
						if Value == key {
							ap = false
							break
						}
					}
					if ap {
						folV = append(folV, key)
						folNCores := w.node.KNodesLookUp(folID)
						w.node.StoreInNodes(folNCores, folID, strings.Join(folV, "!"))
					}
				} else {
					folNCores := w.node.KNodesLookUp(folID)
					w.node.StoreInNodes(folNCores, folID, key)
				}
				c.JSON(http.StatusOK, gin.H{
					"value":  fmt.Sprintf("%s at %s", val, nodestr),
					"folder": folder})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"insertFail": "empty key or value"})
			}

		})

		api.POST("/add", func(c *gin.Context) {
			key := strings.ToUpper(c.PostForm("insertkey"))
			val := c.PostForm("insertval")
			folder := c.PostForm("insertfol")
			if key != "" && val != "" {
				//if no folder listed use EMPTY FOLDER
				if folder == "" || folder == " " {
					folder = "NOFOLDER"
				}
				folder = strings.ToUpper("/" + folder)
				keyID := NewSHA1ID(key)
				folID := NewSHA1ID(folder)
				log.Printf("insert: %s -> %s -> %s of val %s \n", key, keyID, keyID.String(), val)
				log.Printf("insert: %s in folder %s", key, folder)
				nodeCores := w.node.KNodesLookUp(keyID)
				nodeCArray := make([]string, 0)
				nodestr := ""
				for _, nc := range nodeCores {
					nodestr += fmt.Sprintf("%s:%s\t", nc.GUID.String(), nc.IP.String())
					nodeCArray = append(nodeCArray, nc.IP.String())
				}
				log.Printf("Storing at nodecores: %s\n", nodestr)
				w.node.StoreInNodes(nodeCores, keyID, key+"*"+val)
				//find folder values
				folstr := w.node.FindValueByKey(folID)
				if folstr != "value not found" {
					folV := strings.Split(folstr, "!")
					ap := true
					for _, Value := range folV {
						if Value == key {
							ap = false
							break
						}
					}
					if ap {
						folV = append(folV, key)
						folNCores := w.node.KNodesLookUp(folID)
						w.node.StoreInNodes(folNCores, folID, strings.Join(folV, "!"))
					}
				} else {
					folNCores := w.node.KNodesLookUp(folID)
					w.node.StoreInNodes(folNCores, folID, key)
				}
				dataLinks := w.node.getCacheOnlyKeys()
				c.HTML(http.StatusOK, "index.html", gin.H{
					"Title":    "BEEP_BOOP_BM",
					"NodeInfo": w.node.NodeCore.String(),
					"Status":   fmt.Sprintf("Storing %s of Folder %s at %s", val, folder, strings.Join(nodeCArray, " & ")),
					"Payload":  dataLinks,
					"Debug":    false,
				})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"insertFail": "empty key or value"})
			}

		})
		// search: lookup key and return value (need to implement return of string in FindValueByKey)
		api.GET("/readNeighbors", func(c *gin.Context) {
			s := ""
			for i := 0; i < ID_LENGTH*8; i++ {
				bucket := w.node.RoutingTable.Buckets[i]
				for e := bucket.Front(); e != nil; e = e.Next() {
					s += fmt.Sprintf("Bucket %d: %s\n", i, e.Value.(*NodeCore).String())
				}
			}
			log.Printf("readNeighbors: %s\n", s)
			c.JSON(http.StatusOK, gin.H{
				"Search Value By Key": s})
		})
	}

}

func (w WebServer) runWebServer(port int) {

	w.router.LoadHTMLGlob("templates/*")
	w.initializeRoutes()
	w.router.Run(fmt.Sprintf(":%d", port))
}
