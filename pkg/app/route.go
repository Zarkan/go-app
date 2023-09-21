package app

import (
	"regexp"
	"sync"
)

var (
	routes = makeRouter()
)

// Route associates a given path with a function that generates a new Composer
// component. When a user navigates to the specified path, the function
// newComponent is invoked to create and mount the associated component.
//
// Example:
//
//	Route("/home", func() Composer {
//	    return NewHomeComponent()
//	})
func Route(path string, newComponent func() Composer) {
	routes.route(path, newComponent)
}

// RouteWithRegexp associates a URL path pattern with a function that generates
// a new Composer component. When a user navigates to a URL path that matches
// the given regular expression pattern, the function newComponent is invoked to
// create and mount the associated component.
//
// Example:
//
//	RouteWithRegexp("^/users/[0-9]+$", func() Composer {
//	    return NewUserComponent()
//	})
func RouteWithRegexp(pattern string, newComponent func() Composer) {
	routes.routeWithRegexp(pattern, newComponent)
}

type router struct {
	mu               sync.RWMutex
	routes           map[string]func() Composer
	routesWithRegexp []regexpRoute
}

func makeRouter() router {
	return router{
		routes: make(map[string]func() Composer),
	}
}

func (r *router) route(path string, newComponent func() Composer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.routes[path] = newComponent
}

func (r *router) routeWithRegexp(pattern string, newComponent func() Composer) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.routesWithRegexp = append(r.routesWithRegexp, regexpRoute{
		regexp:       regexp.MustCompile(pattern),
		newComponent: newComponent,
	})
}

func (r *router) createComponent(path string) (Composer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	newComponent, isRouted := r.routes[path]
	if !isRouted {
		for _, rwr := range r.routesWithRegexp {
			if rwr.regexp.MatchString(path) {
				newComponent = rwr.newComponent
				isRouted = true
				break
			}
		}
	}
	if !isRouted {
		return nil, false
	}

	return newComponent(), true
}

type regexpRoute struct {
	regexp       *regexp.Regexp
	newComponent func() Composer
}
