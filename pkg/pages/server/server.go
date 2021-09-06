package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/gorilla/mux"
	"github.com/karlseguin/ccache/v2"
	"github.com/linefusion/pages/pkg/pages/config"
	"github.com/linefusion/pages/pkg/pages/sources"
)

type Server struct {
	http    http.Server
	config  config.ServerConfig
	context context.Context
}

type ServerHandler struct {
	handlers *ccache.Cache
	Page     config.PageBlock
	Source   sources.Source
	Server   *Server
}

func NewServerHandler(server *Server, page config.PageBlock, source sources.Source) *ServerHandler {
	handler := &ServerHandler{
		Server: server,
		Page:   page,
		Source: source,
	}
	handler.handlers = ccache.New(ccache.Configure())
	return handler
}

func (handler ServerHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {

	requestKey := handler.Source.CreateKey(request)

	var item *ccache.Item = handler.handlers.Get(requestKey)

	var serve http.Handler

	if item != nil {
		serve = item.Value().(http.Handler)
	} else {
		context := config.CreateRequestContext(request)

		sourceFs, err := handler.Source.CreateFs(context, request)
		if err != nil {
			response.WriteHeader(500)
			response.Write([]byte("500 Internal Server Error"))
			return
		}

		serve = http.FileServer(http.FS(sourceFs))
		handler.handlers.Set(requestKey, serve, 0)
	}

	serve.ServeHTTP(response, request)
}

func New(ctx context.Context, conf config.ServerConfig) Server {
	return Server{
		config:  conf,
		context: ctx,
	}
}

func (server *Server) Start() {
	bind := fmt.Sprintf("%s:%d", server.config.Listen.Bind, server.config.Listen.Port)

	fmt.Printf("Starting server \"%s\" on \"%s\" (%s/%s)\n", server.config.Name, bind, runtime.GOOS, runtime.GOARCH)

	router := mux.NewRouter()

	for _, page := range server.config.Pages.Entries {
		if page.Path == "" {
			page.Path = "/"
		}

		source, err := sources.New(page.Source)
		if err != nil {
			log.Fatal(err)
		}

		if page.Enabled != nil {
			if !*page.Enabled {
				continue
			}
		}

		if len(page.Hosts) > 0 {
			for _, host := range page.Hosts {
				fmt.Printf(" > Exposing page \"%s\" matching host \"%s\" \n", page.Name, host)
				router.
					Host(host).
					Subrouter().
					PathPrefix(page.Path).
					Handler(NewServerHandler(server, page, source))
			}
		} else {
			fmt.Printf(" > Exposing page \"%s\" matching any host\n", page.Name)
			router.
				PathPrefix(page.Path).
				Handler(NewServerHandler(server, page, source))
		}
	}

	server.http = http.Server{
		Addr:         bind,
		Handler:      router,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	go func() {
		if err := server.http.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("%s:ListenAndServe(): %v", server.config.Name, err)
		}
	}()
}

func (server *Server) Stop() {
	server.http.Shutdown(server.context)
}

func (server *Server) Wait() {
	<-server.context.Done()
}
