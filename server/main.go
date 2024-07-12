package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

var cli client.Client
var ctx context.Context

func main() {
	router := gin.Default()
	router.POST("/gandi/collections/create", createCollection)
	router.POST("/gandi/entities/get", getWithID)
	router.POST("/gandi/entities/insert", insertVector)
	router.POST("/gandi/entities/delete", deleteVectors)
	router.POST("/gandi/entities/upsert", upsertVector)

	ctx = context.Background()

	os.Setenv("MILVUS", "localhost:19530")
	var err error
	cli, err = client.NewClient(ctx, client.Config{
		Address: os.Getenv("MILVUS"),
	})
	if err != nil {
		// handling error and exit, to make example simple here
		log.Fatal("failed to connect to milvus:", err.Error())
	}
	defer cli.Close()
	router.Run("localhost:8080")
}

type collectionData struct {
	CollectionName string `json:"collectionName"`
	Dimension      int    `json:"dimension"`
}

func createCollection(c *gin.Context) {
	var newData collectionData

	if err := c.BindJSON(&newData); err != nil {
		fmt.Println("Could not bind data")
		c.JSON(http.StatusNotAcceptable, gin.H{
			"code": http.StatusNotAcceptable,
		})
		return
	}

	schema := entity.NewSchema().WithName(newData.CollectionName).WithDescription("this is the basic example collection").
		WithField(entity.NewField().WithName("id").WithDataType(entity.FieldTypeInt64).WithIsPrimaryKey(true).WithIsAutoID(false)).
		WithField(entity.NewField().WithName("vector").WithDataType(entity.FieldTypeFloatVector).WithDim(int64(newData.Dimension)))

	err := cli.CreateCollection(ctx, schema, entity.DefaultShardNumber)
	if err != nil {
		log.Fatal("failed to create collection:", err.Error())
	}

	idx, err := entity.NewIndexIvfFlat(entity.L2, 2)
	if err != nil {
		log.Fatal("fail to create ivf flat index:", err.Error())
	}
	err = cli.CreateIndex(ctx, newData.CollectionName, "vector", idx, false)
	if err != nil {
		log.Fatal("fail to create index:", err.Error())
	}

	fmt.Print("Collection created")
	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
	})
}

type GetData struct {
	IDs            []int64 `json:"id"`
	CollectionName string  `json:"collectionName"`
}

func getWithID(c *gin.Context) {
	var newData GetData

	if err := c.BindJSON(&newData); err != nil {
		fmt.Println("Could not bind data")
		c.JSON(http.StatusNotAcceptable, gin.H{
			"code": http.StatusNotAcceptable,
		})
		return
	}

	idColumn := entity.NewColumnInt64("id", newData.IDs)

	if err := cli.LoadCollection(ctx, newData.CollectionName, false); err != nil {
		log.Fatal("Could not load collection", err.Error())
	}

	res, err := cli.Get(ctx, newData.CollectionName, idColumn)

	if err != nil {
		log.Fatal("Could not get the data", err.Error())
	}

	fmt.Println(res.GetColumn("id"))
	fmt.Println(res.GetColumn("vector"))

	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": res.GetColumn("vector").FieldData(),
	})
}

type element struct {
	ID     int64     `json:"id"`
	Vector []float32 `json:"vector"`
}

type InsertData struct {
	Data           []element `json:"data"`
	CollectionName string    `json:"collectionName"`
}

func insertVector(c *gin.Context) {
	var newData InsertData

	if err := c.BindJSON(&newData); err != nil {
		fmt.Println("Could not bind data")
		c.JSON(http.StatusNotAcceptable, gin.H{
			"code": http.StatusNotAcceptable,
		})
		return
	}

	IDs := make([]int64, 0, len(newData.Data))
	vecs := make([][]float32, 0, len(newData.Data))

	for _, e := range newData.Data {
		IDs = append(IDs, e.ID)
		vecs = append(vecs, e.Vector)
	}

	idColumn := entity.NewColumnInt64("id", IDs)
	vecColumn := entity.NewColumnFloatVector("vector", 5, vecs)

	fmt.Println(idColumn)
	fmt.Println(vecColumn)

	_, err := cli.DescribeCollection(ctx, newData.CollectionName)

	if err != nil {
		log.Fatal(err.Error())
	}

	ids, err := cli.Insert(ctx, newData.CollectionName, "_default", idColumn, vecColumn)

	if err != nil {
		log.Fatal("Failed to insert vectors", err.Error())
	}

	fmt.Println("IDs of inserted vectors: ", ids)

	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": newData,
	})
}

type DeleteData struct {
	DatabaseName   string
	CollectionName string
	Filter         string
	PartitionName  string
}

func deleteVectors(c *gin.Context) {
	var newData DeleteData

	if err := c.BindJSON(&newData); err != nil {
		fmt.Println("Could not bind data")
		c.JSON(http.StatusNotAcceptable, gin.H{
			"code":    http.StatusNotAcceptable,
			"message": err.Error(),
		})
		return
	}

	if newData.DatabaseName == "" {
		newData.DatabaseName = "default"
	} else if err := cli.UsingDatabase(ctx, newData.DatabaseName); err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{
			"code":    http.StatusNotAcceptable,
			"message": err.Error(),
		})
		fmt.Println("Database does not exist")
	}

	if newData.PartitionName == "" {
		newData.PartitionName = "_default"
	}

	err := cli.Delete(ctx, newData.CollectionName, newData.PartitionName, newData.Filter)

	if err != nil {
		c.JSON(http.StatusNotAcceptable, gin.H{
			"code":    http.StatusNotAcceptable,
			"message": err.Error(),
		})
		fmt.Println("Could not delete specified vectors")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": "",
	})

}

type UpsertData struct {
	Data           []element `json:"data"`
	CollectionName string    `json:"collectionName"`
}

func upsertVector(c *gin.Context) {
	var newData UpsertData

	if err := c.BindJSON(&newData); err != nil {
		fmt.Println("Could not bind data")
		c.JSON(http.StatusNotAcceptable, gin.H{
			"code": http.StatusNotAcceptable,
		})
		return
	}

	IDs := make([]int64, 0, len(newData.Data))
	vecs := make([][]float32, 0, len(newData.Data))

	for _, e := range newData.Data {
		IDs = append(IDs, e.ID)
		vecs = append(vecs, e.Vector)
	}

	idColumn := entity.NewColumnInt64("id", IDs)
	vecColumn := entity.NewColumnFloatVector("vector", 5, vecs)

	ids, err := cli.Upsert(ctx, newData.CollectionName, "_default", idColumn, vecColumn)

	if err != nil {
		log.Fatal("Failed to upsert vectors", err.Error())
	}

	fmt.Println("IDs of upserted vectors: ", ids)

	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": newData,
	})
}
