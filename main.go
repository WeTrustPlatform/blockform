package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"

	"github.com/WeTrustPlatform/blockform/aws"
	"github.com/WeTrustPlatform/blockform/azure"
	"github.com/WeTrustPlatform/blockform/model"
	"github.com/gin-gonic/gin"
	"github.com/sethvargo/go-password/password"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// CloudProvider abstracts the behaviour of a cloud provider like AWS, Azure or
// Google Cloud. It exposes functions to create a virtual machine, install
// an ethereum node on it, and delete a virtual machine.
type CloudProvider interface {
	CreateNode(context.Context, model.Node, func(string, string))
	DeleteNode(context.Context, model.Node, func())
}

var azureProvider CloudProvider
var awsProvider CloudProvider

func providerForNode(node model.Node) CloudProvider {
	var cloud CloudProvider
	switch node.CloudProvider {
	case "aws":
		cloud = awsProvider
	case "azure":
		cloud = azureProvider
	default:
		cloud = awsProvider
	}
	return cloud
}

func main() {
	db, err := gorm.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}

	//db.DropTableIfExists(&model.Node{})
	db.AutoMigrate(&model.Node{})

	azureProvider = azure.NewAzure()
	awsProvider = aws.NewAWS()

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	r.Static("/static", "./static")

	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		os.Getenv("BASIC_AUTH_USER"): os.Getenv("BASIC_AUTH_PASS"),
	}))

	authorized.GET("/", func(c *gin.Context) {
		var nodes []model.Node
		db.Find(&nodes).Order("created_at DESC")
		c.HTML(http.StatusOK, "index.html", nodes)
	})

	authorized.GET("/node/:id", func(c *gin.Context) {
		c.Request.URL.Path = "/node/" + c.Param("id") + "/general"
		r.HandleContext(c)
	})

	authorized.GET("/node/:id/:tab", func(c *gin.Context) {
		id := c.Param("id")
		tab := c.Param("tab")
		if tab == "" {
			tab = "general"
		}
		node := model.Node{}
		db.Find(&node, id)
		c.HTML(http.StatusOK, "node_"+tab+".html", struct {
			Tab  string
			Node model.Node
		}{
			tab,
			node,
		})
	})

	authorized.GET("/create", func(c *gin.Context) {
		c.HTML(http.StatusOK, "create.html", nil)
	})

	authorized.POST("/create", func(c *gin.Context) {
		if err := c.Request.ParseForm(); err != nil {
			c.Status(500)
		}

		name := c.Request.FormValue("name")
		if name == "" {
			name = "bf-" + password.MustGenerate(8, 4, 0, true, false)
		}
		provider := c.Request.FormValue("provider")
		if provider == "" {
			provider = "aws"
		}
		networkID, _ := strconv.ParseInt(c.Request.FormValue("network_id"), 10, 64)
		if networkID == 0 {
			networkID = int64(rand.Intn(1000))
		}
		APIKey := password.MustGenerate(8, 4, 0, false, false)

		node := model.Node{
			Name:          name,
			CloudProvider: provider,
			NetworkType:   c.Request.FormValue("network_type"),
			NetworkID:     uint64(networkID),
			APIKey:        APIKey,
			Status:        model.Creating,
		}

		db.Create(&node)

		cloud := providerForNode(node)
		go cloud.CreateNode(context.Background(), node, func(VMID, DomainName string) {
			db.Model(&node).Update("Status", model.Deployed)
			db.Model(&node).Update("VMID", VMID)
			db.Model(&node).Update("DomainName", DomainName)
			log.Println("Done creating node " + node.Name)
		})

		c.Redirect(http.StatusSeeOther, "/")
	})

	authorized.GET("/delete/:id", func(c *gin.Context) {
		ID := c.Param("id")

		node := model.Node{}
		db.Find(&node, ID)

		db.Model(&model.Node{}).Where("id=?", ID).Update("Status", model.Deleting)

		cloud := providerForNode(node)
		log.Println("Deleting node", node.Name)
		go cloud.DeleteNode(context.Background(), node, func() {
			db.Where("id=?", ID).Delete(&model.Node{})
			log.Println("Done deleting node " + node.Name)
		})

		c.Redirect(http.StatusSeeOther, "/")
	})

	r.Run(":" + os.Getenv("PORT"))
}
