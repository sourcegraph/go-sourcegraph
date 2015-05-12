package sourcegraph

import (
	"sync"

	"google.golang.org/grpc"
)

var (
	connsMu sync.Mutex
	conns   map[string]*grpc.ClientConn // keyed on GRPC target (i.e., addr)
)

// pooledGRPCDial is a global connection pool for grpc.Dial.
func pooledGRPCDial(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	connsMu.Lock()
	if conns == nil {
		conns = map[string]*grpc.ClientConn{}
	}
	if conn := conns[target]; conn != nil {
		connsMu.Unlock()
		return conn, nil
	}
	connsMu.Unlock()

	conn, err := grpc.Dial(target, opts...)
	if err != nil {
		return nil, err
	}

	connsMu.Lock()
	if existing := conns[target]; existing != nil {
		// Another goroutine beat us to establishing the connection;
		// close ours and return the winner's.
		connsMu.Unlock()
		conn.Close()
		return existing, nil
	}
	conns[target] = conn
	connsMu.Unlock()

	return conn, nil
}
