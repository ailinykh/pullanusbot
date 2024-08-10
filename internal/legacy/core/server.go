package core

import "context"

type Server struct {
	Name string
}

type ServerAPI interface {
	GetServers(context.Context) ([]*Server, error)
	RebootServer(context.Context, *Server) error
}
