package handler

import (
	"github.com/xiusin/logger"
	"runtime/debug"
	"sync"
)

type Handler struct {
	sync.Mutex
	routes map[string]HandleFunc
}
type HandleFunc func(data RequestData) string

func (h *Handler) Handle(route string, data RequestData) string {
	h.Lock()
	defer h.Unlock()
	defer func() {
		if err := recover(); err != nil {
			s := debug.Stack()
			logger.Errorf("Recovered Error: %s, ErrorStack: \n%s\n\n", err, string(s))
		}
	}()

	if handle, ok := h.routes[route]; !ok {
		return JSON(ResponseData{Status: FailedCode, Data: nil, Msg: "not found"})
	} else {
		res := handle(data)
		return res
	}
}

func (h *Handler) Add(route string, handle HandleFunc) {
	h.routes[route] = handle
}
