package main

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
)

// Website manage multiple web apps inside to serve web app with cart api
type website struct {
	tracer       opentracing.Tracer
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

	t := setupTracer(siteAppName)
	w := &website{
		engine:       server,
		tracer:       t,
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

	req := c.Request
	ctx, err := w.tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
	var span opentracing.Span
	if err != nil {
		span = w.tracer.StartSpan(buyStuffOperation)
	} else {
		span = w.tracer.StartSpan(buyStuffOperation, opentracing.ChildOf(ctx))
	}
	defer span.Finish()
	ext.SpanKindRPCServer.Set(span)
	span.SetTag(instanceIDTag, w.instanceID)

	if len(sku) < 1 {
		ext.Error.Set(span, true)
		err := errors.New("missing SKU")
		span.LogFields(log.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "no SKUs are present in the request"})
		return
	}

	w.complexBusinessLogic(span)

	if err = w.storeRecordsInRemoteStorage(span); err != nil {
		ext.Error.Set(span, true)
		span.LogFields(log.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "could not store records into the storage"})
		return
	}

	if err := w.addToCart(sku, span); err != nil {
		ext.Error.Set(span, true)
		span.LogFields(log.Error(err))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "could not add the item to the cart"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%s was added to the cart", sku)})
}

func (w *website) addToCart(sku string, parent opentracing.Span) error {
	span := w.tracer.StartSpan(addToCartOperation, opentracing.ChildOf(parent.Context()))
	defer span.Finish()
	ext.SpanKindRPCClient.Set(span)

	url := fmt.Sprintf("http://localhost:8085/cart/%s", sku)
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return err
	}

	err = w.tracer.Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header))
	if err != nil {
		w.log(fmt.Sprintf("could't inject context, continuing without : %v", err))
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

func (w *website) complexBusinessLogic(parent opentracing.Span) {
	span := w.tracer.StartSpan(doComplexBusinessLogicOperation, opentracing.ChildOf(parent.Context()))
	defer span.Finish()
	w.faultManager.sleepForAWhile(siteAppName)
}

func (w *website) storeRecordsInRemoteStorage(parent opentracing.Span) error {
	span := w.tracer.StartSpan(storeRecordsOperation, opentracing.ChildOf(parent.Context()))
	defer span.Finish()
	ext.SpanKindRPCClient.Set(span)

	return w.faultManager.maybeFailTheOperation(siteAppName)
}

func (w *website) log(message string) {
	fmt.Println(message)
}
