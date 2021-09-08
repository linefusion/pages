package server

import (
	"errors"
	"io/fs"

	"github.com/linefusion/pages/pkg/pages/config"
	"github.com/valyala/fasthttp"
)

var ErrPageShadowing = errors.New("fallback page detected in the middle of page chain")

type Router struct {
	pages []Page
}

func (router *Router) handle(context *fasthttp.RequestCtx) {

	// TODO: this lookup and the extracted vars can be cached
	for _, page := range router.pages {
		if matches, vars := page.Matches(context); matches {
			err := page.Serve(context, vars)
			if pathErr, ok := err.(*fs.PathError); ok {
				err = pathErr.Unwrap()
			}

			switch err {
			case fs.ErrNotExist:
				context.Error("404 not found", 404)
				return
			}
			return
		}
	}

	context.Error("unconfigured route", 501)
}

func (router *Router) Build() (fasthttp.RequestHandler, error) {

	fallbackIndex := -1
	for index, route := range router.pages {
		if !route.IsFallback() {
			if fallbackIndex >= 0 {
				return nil, ErrPageShadowing
			}
		} else {
			fallbackIndex = index
		}
	}

	/*
		firstFallbackIndex := -1
		for _, page := range router.routes {
			source, err := sources.New(page.Source)
			if err != nil {
				return nil, err
			}

			if len(page.Hosts) > 0 {
				for _, host := range page.Hosts {
					if firstFallbackIndex >= 0 {
						perror.Printf("✖ page \"%s\" will never be served\n", page.Name)
					} else {
						psuccess.Printf("✔ page \"%s\" responding on \"%s\" \n", page.Name, host)
					}
					router.AddRoute(host, page.Path, NewHandler(server, page, source))
				}
			} else {
				psuccess.Printf("✔ page \"%s\" responding as fallback\n", page.Name)
				router.SetDefaultRoute(page.Path, NewHandler(server, page, source))
				if firstFallbackIndex < 0 {
					firstFallbackIndex = index
				}
			}
		}

		if firstFallbackIndex >= 0 {

			if firstFallbackIndex < len(server.config.Pages.Entries)-1 {
				page := server.config.Pages.Entries[firstFallbackIndex]
				err := fmt.Sprintf("\n"+
					"-----------------------------------------------------------------\n"+
					" WARNING\n"+
					"-----------------------------------------------------------------\n"+
					" Page \"%s\" is being used as a fallback page,\n"+
					" but its not the last page in the serving chain.\n"+
					" Since page blocks are always evaluated top down, the following\n"+
					" pages will never be served:\n\n", page.Name)

				for i := firstFallbackIndex + 1; i < len(server.config.Pages.Entries); i++ {
					err += fmt.Sprintf("  - %s (%s)\n", server.config.Pages.Entries[i].Name, server.config.Pages.Entries[i].Hosts)
				}

				err = err + "\n" +
					" If you want to temporarily disable a page, use\n" +
					" `disable = true` inside that page block instead.\n" +
					"-----------------------------------------------------------------\n"

				pwarning.Print(err)
			}
		}
	*/

	return router.handle, nil
}

func (router *Router) Add(page config.PageBlock) {
	router.pages = append(router.pages, NewRoute(page))
}

func NewRouter() Router {
	return Router{}
}
