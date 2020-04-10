package grpc_pool

import (
	"net"
	"os"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type test struct {
	srvInfo struct {
		srv  *grpc.Server
		Addr string
		IP   net.IP
		Port int
	}
}

var te *test

func TestMain(m *testing.M) {
	te = &test{}
	if err := te.startServer(); err != nil {
		panic(err)
	}
	defer te.stop()
	os.Exit(m.Run())
}

func TestNewGrpcPool(t *testing.T) {
	newClient := func() (*grpc.ClientConn, error) {
		opts := []grpc.DialOption{grpc.WithInsecure(), grpc.WithBlock()}
		return grpc.Dial(te.srvInfo.Addr, opts...)
	}
	pool := NewGrpcPool(newClient, 10, time.Second*1)
	con, err := pool.GetConn()
	if err != nil {
		t.Fatal(err)
	}
	if con.GetState() != connectivity.Ready {
		t.Fatal("client not ready")
	}
	if err := con.Release(); err != nil {
		t.Fatal(err)
	}
	if pool.Len() < 1 {
		t.Fatal("pool len is not right")
	}
	time.Sleep(time.Second)
	con, err = pool.GetConn()
	if err != nil {
		t.Fatal(err)
	}
	if con.GetState() != connectivity.Ready {
		t.Fatal("client not ready")
	}
	con.Release()
	pool.CloseAllConn()

}

func Test_NewGrpcPoolInfinity(t *testing.T) {
	newClient := func() (*grpc.ClientConn, error) {
		opts := []grpc.DialOption{grpc.WithInsecure(), grpc.WithBlock()}
		return grpc.Dial(te.srvInfo.Addr, opts...)
	}
	pool := NewGrpcPool(newClient, 10, -1)
	con, err := pool.GetConn()
	if err != nil {
		t.Fatal(err)
	}
	if con.GetState() != connectivity.Ready {
		t.Fatal("client not ready")
	}
	if err := con.Release(); err != nil {
		t.Fatal(err)
	}
	if pool.Len() < 1 {
		t.Fatal("pool len is not right")
	}
	time.Sleep(time.Second)
	con, err = pool.GetConn()
	if err != nil {
		t.Fatal(err)
	}
	if con.GetState() != connectivity.Ready {
		t.Fatal("client not ready")
	}
	con.Release()
	pool.CloseAllConn()

}

func TestNewGrpcPool2(t *testing.T) {
	newClient := func() (*grpc.ClientConn, error) {
		opts := []grpc.DialOption{grpc.WithInsecure(), grpc.WithBlock()}
		return grpc.Dial(te.srvInfo.Addr, opts...)
	}
	pool := NewGrpcPool(newClient, 5, 1)
	getConn := func() {
		con, err := pool.GetConn()
		if err != nil {
			t.Fatal(err)
		}
		if con.GetState() != connectivity.Ready {
			t.Fatal("client not ready")
		}
		time.AfterFunc(time.Second/100, func() {
			if err := con.Release(); err != nil {
				t.Fatal(err)
			}
		})
	}
	for i := 0; i < 20; i++ {
		t.Logf("index: %d", i)
		t.Logf("current len of pool: %d\n", pool.Len())
		getConn()
	}
	t.Logf("current len of pool: %d\n", pool.Len())
	time.Sleep(time.Second * 2)
	t.Logf("current len of pool: %d\n", pool.Len())
	getConn()
	getConn()
	getConn()
	t.Logf("current len of pool: %d\n", pool.Len())
	time.Sleep(time.Second * 2)
	t.Logf("current len of pool: %d\n", pool.Len())
	pool.CloseAllConn()
}

func (te *test) startServer() error {
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return err
	}
	server := grpc.NewServer()
	te.srvInfo.srv = server
	go server.Serve(lis)
	te.srvInfo.Addr = lis.Addr().String()
	te.srvInfo.IP = lis.Addr().(*net.TCPAddr).IP
	te.srvInfo.Port = lis.Addr().(*net.TCPAddr).Port
	return nil
}

func (te *test) stop() {
	te.srvInfo.srv.Stop()
}

func TestNewGrpcPool3(t *testing.T) {
	createdTotal := 0
	newClient := func() (*grpc.ClientConn, error) {
		opts := []grpc.DialOption{grpc.WithInsecure(), grpc.WithBlock()}
		createdTotal++
		return grpc.Dial(te.srvInfo.Addr, opts...)
	}
	pool := NewGrpcPool(newClient, 5, -1)
	wg := sync.WaitGroup{}
	total := 100
	wg.Add(total)
	getConn := func() {
		t.Logf("current len of pool: %d\n", pool.Len())
		con, err := pool.GetConn()
		if err != nil {
			t.Fatal(err)
		}
		con.Release()
		t.Logf("release len of pool: %d\n", pool.Len())
	}

	for i := 0; i < total; i++ {
		go func() {
			getConn()
			wg.Done()
		}()
	}
	wg.Wait()
	t.Log("created total: ", createdTotal)
	for i := 0; i < 10; i++ {
		getConn()

	}
	t.Logf("current len of pool: %d\n", pool.Len())
	pool.CloseAllConn()
}