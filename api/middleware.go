package api

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ron96G/whatsapp-bizapi-mock/util"
	"github.com/uber/jaeger-client-go/config"
	"github.com/valyala/fasthttp"
	"golang.org/x/time/rate"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	log "github.com/ron96G/go-common-utils/log"
)

const (
	serviceName     = "wabiz-mockserver"
	componentName   = "fasthttp"
	requestIDHeader = "X-Request-ID"
)

func (a *API) Authorize(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		logger := a.LoggerFromCtx(ctx)
		token, _ := extractAuthToken(ctx, "Bearer")

		if a.APIKey != "" && token == a.APIKey {
			logger.Info("Successfully authorized user with API key")
			h(ctx)
			return
		}

		logger.Warn("Failed to authorize user", "reason", "invalid API key")
		ctx.SetStatusCode(401)
	})
}

func AuthorizeStaticToken(h fasthttp.RequestHandler, staticToken string) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		logger := log.FromContext(ctx)
		token, _ := extractAuthToken(ctx, "Apikey")
		if staticToken != "" && token != staticToken {
			logger.Warn("Failed to authorize user", "reason", "invalid apikey")
			ctx.SetStatusCode(401)
			return
		}

		logger.Info("Successfully authorized user with apikey")
		h(ctx)
	})
}

func Limiter(h fasthttp.RequestHandler, concurrencyLimit uint) fasthttp.RequestHandler {
	limiter := rate.NewLimiter(rate.Limit(concurrencyLimit), int(concurrencyLimit))
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {
		if !limiter.Allow() {
			ctx.SetStatusCode(429)
			return
		}
		h(ctx)
	})
}

func (a *API) SetConnID(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {

		reqID := ctx.Request.Header.Peek(requestIDHeader)

		if len(reqID) == 0 {
			reqID = []byte(uuid.New().String())
			ctx.Request.Header.SetBytesV(requestIDHeader, reqID)
			ctx.Response.Header.SetBytesV(requestIDHeader, reqID)

		} else {
			ctx.Response.Header.SetBytesV(requestIDHeader, reqID)
		}
		logger := a.LoggerFromCtx(ctx).New("id", string(reqID))
		LoggerToCtx(ctx, logger)
		h(ctx)
	})
}

func Tracer(h fasthttp.RequestHandler) fasthttp.RequestHandler {

	defcfg := config.Configuration{
		ServiceName: serviceName,
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
		},
	}

	config, err := defcfg.FromEnv()
	if err != nil {
		panic("Could not parse Jaeger env vars: " + err.Error())
	}

	tr, _, err := config.NewTracer()
	if err != nil {
		panic("Could not initialize jaeger tracer: " + err.Error())
	}

	opentracing.SetGlobalTracer(tr)

	return fasthttp.RequestHandler(func(ctx *fasthttp.RequestCtx) {

		req := ctx.Request
		method := string(ctx.Method())
		url := string(ctx.Path())
		opname := "HTTP " + method + " URL: " + url
		var sp opentracing.Span
		carrier := util.NewCarrier(&req.Header)

		if c, err := tr.Extract(opentracing.HTTPHeaders, carrier); err != nil {
			sp = tr.StartSpan(opname)
		} else {
			sp = tr.StartSpan(opname, ext.RPCServerOption(c))
		}

		ext.HTTPMethod.Set(sp, method)
		ext.HTTPUrl.Set(sp, url)
		ext.Component.Set(sp, componentName)

		ctx.SetUserValue("activeSpan", sp)

		h(ctx)
		status := uint16(ctx.Response.StatusCode())
		ext.HTTPStatusCode.Set(sp, status)

		if status >= http.StatusInternalServerError {
			ext.Error.Set(sp, true)
		}

		sp.Finish()
	})
}
