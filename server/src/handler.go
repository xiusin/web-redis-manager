package src

import "github.com/asticode/go-astilog"

type Handler struct {
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
      astilog.Errorf("Recovered Error: %s", err)
    }
  }()
  handle, ok := h.routes[route]
  if !ok {
    return JSON(ResponseData{Status: 5000, Data: nil, Msg: "notfound"})
  }
  res := handle(data)
  return res
}

func (h *Handler) Add(route string, handle HandleFunc) *Handler {
  h.routes[route] = handle
  return h
}
