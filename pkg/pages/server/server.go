package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/gorilla/mux"
	"github.com/linefusion/pages/pkg/pages/config"
	"github.com/linefusion/pages/pkg/pages/sources"
	"github.com/spf13/afero"
)

type Server struct {
	http    http.Server
	config  config.ServerConfig
	context context.Context
}

type ServerHandler struct {
	Page   config.PageBlock
	Source sources.Source
}

func (handler ServerHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	fs, err := handler.Source.Fs(request)
	if err != nil {
		response.WriteHeader(500)
		response.Write([]byte(""))
		return
	}

	httpFs := afero.NewHttpFs(fs)
	requestHandler := http.FileServer(httpFs.Dir("/"))
	requestHandler.ServeHTTP(response, request)
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
					Handler(&ServerHandler{
						Page:   page,
						Source: source,
					})
			}
		} else {
			fmt.Printf(" > Exposing page \"%s\" matching any host\n", page.Name)
			router.
				PathPrefix(page.Path).
				Handler(&ServerHandler{
					Page:   page,
					Source: source,
				})
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
