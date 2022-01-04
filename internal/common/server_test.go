package common

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
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
	STEP: []string{STEPOK_default},
}

func NewTcpServer(ctx context.Context, t testing.TB, r respMap) *TcpServer {
	svr := new(TcpServer)
	svr.ctx = ctx
	svr.rMap = r
	svr.printConnErr = true

	l, err := net.Listen("tcp", "127.0.0.1:0")

	if err != nil {
		t.Fatalf("Caught an error and didn't expect one: %v", err)
	}

	svr.listener = l
	svr.addr = svr.listener.Addr().String()

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
	ctx          context.Context
	addr         string
	Host         string
	Port         int
	listener     net.Listener
	conn         net.Conn
	rMap         respMap
	printConnErr bool
}

func (t *TcpServer) stop() {
	select {
	case <-t.ctx.Done():
		t.Close()
	}
	return
}

func (t *TcpServer) Start(wg *sync.WaitGroup) (err error) {
	var reqType ServerMessageType
	wg.Done()
	respCount := 0
	for {
		select {
		case <-t.ctx.Done():
			return nil
		default:
		}
		// logger.Debug("Waiting for new connection")
		t.conn, err = t.listener.Accept()
		if err != nil {
			err = errors.New("could not accept connection")
			break
		}
		if t.conn == nil {
			err = errors.New("could not create connection")
			break
		}
		// logger.Debug("Got new connection")

		scanner := bufio.NewScanner(t.conn)

		for scanner.Scan() {
			// TODO: This mod scheme wont work if I have more than one rMap configured
			respCount++
			req := scanner.Text()
			// logger.Debug("Svr conn output: ", req)
			// Strip auth prefix from command: "poolPw walletAddr Command"
			reqSplit := strings.SplitN(req, " ", 3)
			if len(reqSplit) < 3 {
				logger.Debug("wth is this? ", req)
				continue
			}
			req = reqSplit[2]
			reqType, err = getReqType(req)
			if err != nil {
				panic(err)
			}

			// TODO: need to pop and/or cycle through slice
			var idx int
			resp, ok := t.rMap[reqType]
			if ok {
				idx = (len(resp) - 1) % respCount
			} else {
				resp, ok = defaultRespMap[reqType]
				if !ok {
					pMsg := `Could not find a response for request in rMap
Req:  %s
rMap: %v`
					panic(fmt.Sprintf(pMsg, req, t.rMap))
				}
				idx = 0
			}

			fmt.Fprintln(t.conn, resp[idx])
		}

		if err := scanner.Err(); err != nil && t.printConnErr {
			logger.Debugf("error reading connection: ", err)
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
