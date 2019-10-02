package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
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

	req := c.Request
	ctx, err := api.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	var span opentracing.Span
	if err != nil {
		span = api.tracer.StartSpan(addToCartOperation)
	} else {
		span = api.tracer.StartSpan(addToCartOperation, opentracing.ChildOf(ctx))
	}
	defer span.Finish()
	ext.SpanKindRPCServer.Set(span)
	span.SetTag(instanceIDTag, api.instanceID)
	span.LogFields(log.String("sku", sku))

	if len(sku) == 0 {
		span.SetTag("error", true)
		span.LogKV("message", "missing sku")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing sku"})
		return
	}

	api.complexBusinessLogic(span)

	err = api.storeRecordsInRemoteStorage(span)
	if err != nil {
		ext.Error.Set(span, true)
		span.LogKV("message", fmt.Sprintf("%v", err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "we were not able to process your request"})
		return
	}
	isInStock, err := api.checkStock(sku, span)
	if err != nil {
		ext.Error.Set(span, true)
		span.LogKV("message", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !isInStock {
		c.JSON(http.StatusNotFound, gin.H{"error": "not enough stock"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"success": true})
}

func (api *cartAPI) checkStock(sku string, parent opentracing.Span) (bool, error) {
	span := api.tracer.StartSpan(checkStockOperation, opentracing.ChildOf(parent.Context()))
	ext.SpanKindRPCClient.Set(span)
	defer span.Finish()

	url := fmt.Sprintf(stockAPIEndpoint, sku)
	httpReq, _ := http.NewRequest(http.MethodGet, url, nil)

	resp, err := http.DefaultClient.Do(httpReq) // check stock
	if err != nil {
		ext.Error.Set(span, true)
		span.LogKV("message", err)
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		buf, _ := ioutil.ReadAll(resp.Body)
		ext.Error.Set(span, true)
		message := string(buf)
		span.LogKV("message", message)
		return false, errors.New(message)
	}
	var stock map[string]int
	if err := json.NewDecoder(resp.Body).Decode(&stock); err != nil {
		ext.Error.Set(span, true)
		span.LogKV("message", err)
		return false, err
	}

	if qty, has := stock[sku]; !has || qty < 1 {
		ext.Error.Set(span, true)
		span.LogKV("message", "not enough stock")
		return false, nil
	}

	return true, nil
}

func (api *cartAPI) complexBusinessLogic(parent opentracing.Span) {
	span := api.tracer.StartSpan(doComplexBusinessLogicOperation, opentracing.ChildOf(parent.Context()))
	defer span.Finish()

	api.faultManager.sleepForAWhile(cartAppName)
}

func (api *cartAPI) storeRecordsInRemoteStorage(parent opentracing.Span) error {
	span := api.tracer.StartSpan(storeRecordsOperation, opentracing.ChildOf(parent.Context()))
	ext.SpanKindRPCClient.Set(span)
	defer span.Finish()

	return api.faultManager.maybeFailTheOperation(cartAppName)
}

func (api *cartAPI) log(message string) {
	fmt.Println(message)
}
