package httphandler

import (
	"github.com/valyala/fasthttp"

	"../servertemplate"
)

func DetailHandler(ctx *fasthttp.RequestCtx) {
	p := &servertemplate.DetailPage{
		CTX: ctx,
	}
	ctx.SetContentType("text/html; charset=utf-8")
	servertemplate.WritePageTemplate(ctx, p)
}

func IndexHandler(ctx *fasthttp.RequestCtx) {
	p := &servertemplate.IndexPage{
		CTX: ctx,
	}
	ctx.SetContentType("text/html; charset=utf-8")
	servertemplate.WritePageTemplate(ctx, p)
}
