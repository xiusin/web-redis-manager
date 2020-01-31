package src

import (
  "github.com/asticode/go-astilog"
  "runtime/debug"
  "sync"
)

type Handler struct {
  sync.Mutex
  routes map[string]HandleFunc
}
type HandleFunc func(data map[string]interface{}) string

func NewHandler() *Handler {
  return &Handler{
    routes: make(map[string]HandleFunc),
  }
}
func (h *Handler) Handle(route string, data map[string]interface{}) string {
  defer func() {
    if err := recover(); err != nil {
      s := debug.Stack()
      astilog.Errorf("Recovered Error: %s, ErrorStack: \n%s\n\n", err, string(s))
    }
  }()
  h.Lock()
  defer h.Unlock()
  handle, ok := h.routes[route]
  if !ok {
    return JSON(ResponseData{Status: 5000, Data: nil, Msg: "notfound"})
  }
  res := handle(data)
  return res
}

func (h *Handler) Add(route string, handle HandleFunc) {
  h.Lock()
  defer h.Unlock()
  h.routes[route] = handle
}
