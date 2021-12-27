package common

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
)

type (
	respMap map[ServerMessageType][]string
)

var defaultRespMap = respMap{
	JOIN: []string{JOINOK_default},
	PING: []string{PONG_default},
}

func NewTcpServer(done chan struct{}, t *testing.T, r respMap) *TcpServer {
	svr := new(TcpServer)
	svr.rMap = r

	l, err := net.Listen("tcp", "127.0.0.1:0")

	if err != nil {
		t.Fatalf("Caught an error and didn't expect one: %v", err)
	}

	svr.listener = l
	svr.addr = svr.listener.Addr().String()
	svr.done = done

	host, port, err := net.SplitHostPort(svr.addr)

	if err != nil {
		t.Fatalf("Caught an error and didn't expect one: %v", err)
	}

	svr.Host = host
	svr.Port, _ = strconv.Atoi(port)

	var wg sync.WaitGroup
	wg.Add(1)
	go svr.Start(&wg)
	go svr.stop()

	wg.Wait()

	return svr
}

type TcpServer struct {
	addr     string
	Host     string
	Port     int
	done     chan struct{}
	listener net.Listener
	rMap     respMap
}

func (t *TcpServer) stop() {
	select {
	case <-t.done:
		t.Close()
	}
	return
}

func (t *TcpServer) Start(wg *sync.WaitGroup) (err error) {
	var reqType ServerMessageType
	wg.Done()
	// TODO: need to incorporate either a done channel or context
	for {

		conn, err := t.listener.Accept()
		if err != nil {
			err = errors.New("could not accept connection")
			break
		}
		if conn == nil {
			err = errors.New("could not create connection")
			break
		}

		scanner := bufio.NewScanner(conn)

		for scanner.Scan() {
			req := scanner.Text()
			// fmt.Println("Svr conn output: ", req)
			reqType, err = getReqType(req)
			if err != nil {
				panic(err)
			}

			// // TODO: need to pop and/or cycle through slice
			resp, ok := t.rMap[reqType]
			if !ok {
				resp, ok = defaultRespMap[reqType]
				if !ok {
					pMsg := `Could not find a response for request in rMap
Req:  %s
rMap: %v`
					panic(fmt.Sprintf(pMsg, req, t.rMap))
				}
			}
			fmt.Fprintln(conn, resp[0])
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "error reading connection: ", err)
		}

	}

	return
}

func (t *TcpServer) Close() (err error) {
	return t.listener.Close()
}

func getReqType(s string) (ServerMessageType, error) {
	msg := strings.SplitN(s, " ", 2)[0]
	return stringToType(msg)
}
