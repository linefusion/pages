package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/linefusion/pages/pkg/pages/config"
	"github.com/valyala/fasthttp"
)

type Server struct {
	transport fasthttp.Server
	router    Router
	config    config.ServerConfig
}

func New(conf config.ServerConfig) Server {
	server := Server{
		config: conf,
	}

	server.router = NewRouter()
	for _, page := range server.config.Pages.Entries {
		if page.Enabled != nil {
			if !*page.Enabled {
				continue
			}
		}
		server.router.Add(page)
	}

	return server
}

func (server *Server) Start() error {
	bind := fmt.Sprintf("%s:%d", server.config.Listen.Bind, server.config.Listen.Port)

	handler, err := server.router.Build()
	if err != nil {
		return err
	}

	server.transport = fasthttp.Server{
		Name:                         "linefusion/pages (" + server.config.Name + ")",
		Handler:                      handler,
		WriteTimeout:                 time.Duration(config.DefaultInt(server.config.Options.WriteTimeout, 60)) * time.Second,
		ReadTimeout:                  time.Duration(config.DefaultInt(server.config.Options.ReadTimeout, 60)) * time.Second,
		IdleTimeout:                  time.Duration(config.DefaultInt(server.config.Options.IdleTimeout, 15)) * time.Second,
		Concurrency:                  config.DefaultInt(server.config.Options.Concurrency, fasthttp.DefaultConcurrency),
		MaxConnsPerIP:                config.DefaultInt(server.config.Options.MaxConnsPerIP, 0),
		MaxRequestsPerConn:           config.DefaultInt(server.config.Options.MaxRequestsPerConn, 0),
		ReadBufferSize:               config.DefaultInt(server.config.Options.ReadBufferSize, 4096),
		WriteBufferSize:              config.DefaultInt(server.config.Options.WriteBufferSize, 4096),
		MaxRequestBodySize:           config.DefaultInt(server.config.Options.MaxRequestBodySize, fasthttp.DefaultMaxRequestBodySize),
		ReduceMemoryUsage:            config.DefaultBool(server.config.Options.ReduceMemoryUsage, false),
		TCPKeepalive:                 config.DefaultBool(server.config.Options.TCPKeepalive, false),
		TCPKeepalivePeriod:           time.Duration(config.DefaultInt(server.config.Options.TCPKeepalivePeriod, 0)),
		GetOnly:                      config.DefaultBool(server.config.Options.GetOnly, false),
		DisablePreParseMultipartForm: config.DefaultBool(server.config.Options.DisablePreParseMultipartForm, false),

		DisableKeepalive: config.DefaultBool(server.config.Options.DisableKeepalive, false),
		//MaxKeepaliveDuration: config.DefaultInt(server.config.Options.MaxKeepaliveDuration, time), // deprecated by fasthttp
		// ErrorHandler: config.DefaultFunc(server.config.Options.ErrorHandler, func) // (ctx *RequestCtx, err err,
		// HeaderReceived: config.DefaultFunc(server.config.Options.HeaderReceived, func) // (header *RequestHeader) RequestCon,
		// ContinueHandler: config.DefaultFunc(server.config.Options.ContinueHandler, func) // (header *RequestHeader) b,
		// LogAllErrors: config.DefaultBool(server.config.Options.LogAllErrors, bool),
		// SecureErrorLogMessage: config.DefaultBool(server.config.Options.SecureErrorLogMessage, bool),
		// DisableHeaderNamesNormalizing: config.DefaultBool(server.config.Options.DisableHeaderNamesNormalizing, bool),
		// SleepWhenConcurrencyLimitsExceeded: config.DefaultTime(server.config.Options.SleepWhenConcurrencyLimitsExceeded, time) // .Durat,
		// NoDefaultServerHeader: config.DefaultBool(server.config.Options.NoDefaultServerHeader, bool),
		// NoDefaultDate: config.DefaultBool(server.config.Options.NoDefaultDate, bool),
		// NoDefaultContentType: config.DefaultBool(server.config.Options.NoDefaultContentType, bool),
		// ConnState: config.DefaultFunc(server.config.Options.ConnState, func) // (net.Conn, ConnSta,
		// Logger: config.DefaultLogger(server.config.Options.Logger, Logger),
		// KeepHijackedConns: config.DefaultBool(server.config.Options.KeepHijackedConns, bool),
		// CloseOnShutdown: config.DefaultBool(server.config.Options.CloseOnShutdown, bool),
		// StreamRequestBody: config.DefaultBool(server.config.Options.StreamRequestBody, bool),

	}

	go func() {
		if err := server.transport.ListenAndServe(bind); err != http.ErrServerClosed {
			log.Fatalf("%s:ListenAndServe(): %v", server.config.Name, err)
		}
	}()

	return nil
}

func (server *Server) Stop() {
	server.transport.Shutdown()
}
