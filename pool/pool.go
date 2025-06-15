package pool

import "request_balancer/types"

type ServerPool struct {
	Servers []*types.Server
}

func NewServerPool() *ServerPool {
	return &ServerPool{
		Servers: make([]*types.Server, 0),
	}
}

func (p *ServerPool) AddServer(server *types.Server) error {
	p.Servers = append(p.Servers, server)
	return nil
}

func (p *ServerPool) GetAllServers() []*types.Server {
	return p.Servers
}
