package tree

import (
	"github.com/AnTengye/lol-shield/internal/pkg/syslog"
)

type Engine struct {
	router *Router
}

func NewEngine() *Engine {
	return &Engine{router: NewRouter()}
}

func (e *Engine) RegisterRoute(method string, pattern string, handler HandlerFunc) {
	e.router.AddRoute(method, pattern, handler)
}

func (e *Engine) GetRoute(c *Context) error {
	n, m := e.router.GetRoute(c.Method, c.Path)
	if n != nil {
		c.Params = m
		key := c.Method + "-" + n.Pattern
		return e.router.GetHandler(key)(c)
	} else {
		return e.defaultRouter(c)
	}
}

func (e *Engine) defaultRouter(c *Context) error {
	syslog.L.Debugf("default router,method:%s,uri: %s, params:%+v", c.Method, c.Path, c.Params)
	return nil
}
