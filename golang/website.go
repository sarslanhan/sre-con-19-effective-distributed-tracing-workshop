package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// Website manage multiple web apps inside to serve web app with cart api
type website struct {
	engine       *gin.Engine
	faultManager *FaultInjectionManager
	az           int
	instanceID   string
}

const (
	siteAppName = "super-website"
)

func newWebsite(faultManager *FaultInjectionManager, instanceID string, az int) *website {
	server := gin.Default()
	server.Use(gin.Logger(), gin.Recovery())
	server.LoadHTMLFiles("templates/index.html")
	server.StaticFile("/magic.js", "templates/magic.js")

	w := &website{
		engine:       server,
		faultManager: faultManager,
		az:           az,
		instanceID:   instanceID,
	}
	server.GET("/", w.renderPage)
	server.POST("/buyStuff", w.buyStuff)
	return w
}

// Run the website for processing requests
func (w *website) Run(addr ...string) error {
	return w.engine.Run(addr...)
}

func (w *website) renderPage(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{"sku": ""})
}

func (w *website) buyStuff(c *gin.Context) {
	sku := c.PostForm("sku")

	if len(sku) < 1 {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "no SKUs are present in the request"})
		return
	}

	w.complexBusinessLogic()

	if err := w.storeRecordsInRemoteStorage(); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "could not store records into the storage"})
		return
	}

	if err := w.addToCart(sku); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "could not add the item to the cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%s was added to the cart", sku)})
}

func (w *website) addToCart(sku string) error {
	url := fmt.Sprintf("http://localhost:8085/cart/%s", sku)
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= http.StatusBadRequest {
		buf, _ := ioutil.ReadAll(res.Body)
		msg := string(buf)
		return errors.New(msg)
	}

	return nil
}

func (w *website) complexBusinessLogic() {
	w.faultManager.sleepForAWhile(siteAppName)
}

func (w *website) storeRecordsInRemoteStorage() error {
	return w.faultManager.maybeFailTheOperation(siteAppName)
}

func (w *website) log(message string) {
	fmt.Println(message)
}
