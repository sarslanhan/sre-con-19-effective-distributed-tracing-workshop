package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	opentracing "github.com/opentracing/opentracing-go"
)

const (
	stockAPIEndpoint = "https://sre-con-stock-api.stups-test.zalan.do/stock/%s"
	cartAppName      = "cart-api"
)

type cartAPI struct {
	tracer       opentracing.Tracer
	engine       *gin.Engine
	faultManager *FaultInjectionManager
	az           int
	instanceID   string
}

func newCartAPI(faultManager *FaultInjectionManager, instanceID string, az int) *cartAPI {
	engine := gin.Default()
	engine.Use(gin.Logger(), gin.Recovery())
	api := &cartAPI{
		engine:       engine,
		tracer:       setupTracer(cartAppName),
		faultManager: faultManager,
		az:           az,
		instanceID:   instanceID,
	}
	engine.PUT("/cart/:sku", api.addToCart)
	return api
}

func (api *cartAPI) Run(addr ...string) error {
	return api.engine.Run(addr...)
}

func (api *cartAPI) addToCart(c *gin.Context) {
	sku := c.Param("sku")

	if len(sku) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing sku"})
		return
	}

	api.complexBusinessLogic()

	err := api.storeRecordsInRemoteStorage()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "we were not able to process your request"})
		return
	}
	isInStock, err := api.checkStock(sku)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !isInStock {
		c.JSON(http.StatusNotFound, gin.H{"error": "not enough stock"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"success": true})
}

func (api *cartAPI) checkStock(sku string) (bool, error) {
	url := fmt.Sprintf(stockAPIEndpoint, sku)
	httpReq, _ := http.NewRequest(http.MethodGet, url, nil)

	resp, err := http.DefaultClient.Do(httpReq) // check stock
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		buf, _ := ioutil.ReadAll(resp.Body)
		message := string(buf)
		return false, errors.New(message)
	}
	var stock map[string]int
	if err := json.NewDecoder(resp.Body).Decode(&stock); err != nil {
		return false, err
	}

	if qty, has := stock[sku]; !has || qty < 1 {
		return false, nil
	}

	return true, nil
}

func (api *cartAPI) complexBusinessLogic() {
	api.faultManager.sleepForAWhile(cartAppName)
}

func (api *cartAPI) storeRecordsInRemoteStorage() error {
	return api.faultManager.maybeFailTheOperation(cartAppName)
}

func (api *cartAPI) log(message string) {
	fmt.Println(message)
}
