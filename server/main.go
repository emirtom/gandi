package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.POST("/gandi/collections/create", createCollection)
	router.POST("/gandi/entities/get", getWithID)
	router.POST("/gandi/entities/insert", insertVector)

	router.Run("localhost:19530")
}

func createCollection(c *gin.Context) {
	fmt.Print("Collection created")
	c.IndentedJSON(http.StatusOK, 0)
}

func getWithID(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, 0)
}

type element struct {
	ID     int
	Vector string
}

type InsertData struct {
	Data           []element `json:"data"`
	CollectionName string    `json:"collectionName"`
}

func insertVector(c *gin.Context) {
	var newData InsertData

	if err := c.BindJSON(&newData); err != nil {
		fmt.Println("Could not bind data")
		c.IndentedJSON(http.StatusNotAcceptable, 0)
		return
	}

	// Do something with the data

	c.IndentedJSON(http.StatusOK, newData)
}
