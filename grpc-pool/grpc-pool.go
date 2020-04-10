package grpc_pool

import (
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type grpcPool struct {
	size          int
	clientConnTtl int64
	newGrpcClient NewGrpcClient
	sync.Mutex
	conns []*clientConn
	_     struct{}
}

type clientConn struct {
	*grpc.ClientConn
	pool        *grpcPool
	createdTime int64
	_           struct{}
}

type GrpcPool = *grpcPool
type ClientConn = *clientConn

type NewGrpcClient func() (*grpc.ClientConn, error)

func NewGrpcPool(newConn NewGrpcClient, size int, clientConnTtl time.Duration) GrpcPool {
	if newConn == nil {
		panic("NewGrpcClient func is nil")
	}
	if size < 1 {
		size = 1
	}
	if clientConnTtl == 0 {
		clientConnTtl = time.Second * 30
	}
	return &grpcPool{
		newGrpcClient: newConn,
		size:          size,
		clientConnTtl: int64(clientConnTtl.Seconds()),
		conns:         make([]*clientConn, 0),
	}
}

func (p *grpcPool) GetConn() (ClientConn, error) {
	p.Lock()
	conns := p.conns
	tn := time.Now().Unix()

	for len(conns) > 0 {
		conn := conns[len(conns)-1]
		conns = conns[0 : len(conns)-1]
		p.conns = conns
		if p.clientConnTtl > 0 && (tn-conn.createdTime) > p.clientConnTtl {
			conn.ClientConn.Close()
			conn.ClientConn = nil
			continue
		}
		if conn.ClientConn == nil {
			continue
		}
		if conn.ClientConn.GetState() == connectivity.Shutdown {
			conn.ClientConn.Close()
			conn.ClientConn = nil
			continue
		}
		p.Unlock()
		return conn, nil
	}
	p.Unlock()

	conn, err := p.newGrpcClient()
	if err != nil {
		return nil, err
	}
	return &clientConn{
		ClientConn:  conn,
		pool:        p,
		createdTime: time.Now().Unix(),
	}, nil
}

func (p *grpcPool) CloseAllConn() error {
	p.Lock()
	defer p.Unlock()
	for _, v := range p.conns {
		v.ClientConn.Close()
		v.ClientConn = nil
	}
	p.conns = p.conns[:0]
	return nil
}

func (p *grpcPool) Len() int {
	return len(p.conns)
}

func (c *clientConn) Close() error {
	c.pool.Lock()
	if c.ClientConn == nil {
		return nil
	}
	if len(c.pool.conns) >= c.pool.size {
		c.pool.Unlock()
		err := c.ClientConn.Close()
		c.ClientConn = nil
		return err
	}
	c.pool.conns = append(c.pool.conns, c)
	c.pool.Unlock()
	return nil
}

func (c *clientConn) Release() error {
	return c.Close()
}