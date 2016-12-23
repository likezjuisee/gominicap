package main

import (
	"fmt"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"

	"./httphandler"
	"./websockethandler"
	"adbutil"
)

var (
	Port   = 8080
	router *fasthttprouter.Router
)

func RouterHandler(ctx *fasthttp.RequestCtx) {
	router.Handler(ctx)
}

func main() {
	defer adbutil.Logger.Close()
	router = fasthttprouter.New()
	router.GET("/minicap", websockethandler.MinicapHandler)
	router.GET("/detail", httphandler.DetailHandler)
	router.GET("/", httphandler.IndexHandler)

	port := fmt.Sprintf(":%d", Port)
	fmt.Println("start server ", port)
	fasthttp.ListenAndServe(port, RouterHandler)
}
