package main

import (
	"bufio"
	"bytes"
	"http-article-proxy/article"
	"http-article-proxy/data"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
)

type Server struct {
	port          int
	dest          string
	ConnectionMap sync.Map
}
type ServerConnection struct {
	uuid      string
	conn      net.Conn
	readBuf   bytes.Buffer
	readBufMu sync.Mutex
	closed    bool
	closeChan chan struct{}
}

func NewServer(port int, dest string) *Server {
	return &Server{
		ConnectionMap: sync.Map{},
		dest:          dest,
		port:          port,
	}
}
func (s *Server) Serve() {
	err := http.ListenAndServe(":"+strconv.Itoa(s.port), http.HandlerFunc(s.httpHandler))
	if err != nil {
		panic(err)
	}
}
func (s *Server) httpHandler(writer http.ResponseWriter, request *http.Request) {
	var bodyBuf bytes.Buffer
	_, err := io.Copy(&bodyBuf, request.Body)
	if err != nil {
		log.Println("cannot read body: " + err.Error())
		return
	}
	request.Body.Close()
	uuid := request.RequestURI[1:]
	log.Println("handle request from http: " + uuid)
	connection, connectionExist := s.ConnectionMap.Load(uuid)
	if !connectionExist {
		conn, err := net.Dial("tcp", s.dest)
		if err != nil {
			log.Println("cannot dial: " + err.Error())
			return
		}
		serverConnection := NewServerConnection(conn, uuid)
		go serverConnection.Handle()
		s.ConnectionMap.Store(uuid, serverConnection)
		connection = serverConnection
	}
	serverConnection := connection.(*ServerConnection)
	if bodyBuf.Len() > 0 {
		go serverConnection.Send(bodyBuf.String())
	}
	v := serverConnection.GetBodyToReadAndReset()
	if len(v) == 0 && serverConnection.closed {
		log.Println("write X-Connection == closed")
		writer.Header().Set("X-Connection", "closed")
	}
	writer.WriteHeader(200)
	_, err = writer.Write([]byte(v))
	if err != nil {
		log.Println("cannot write to http writer: " + err.Error())
	}
	//log.Println("write to http: " + string(v))
}
func NewServerConnection(conn net.Conn, uuid string) *ServerConnection {
	log.Println("new connection: " + uuid)
	return &ServerConnection{
		uuid:      uuid,
		conn:      conn,
		readBuf:   bytes.Buffer{},
		readBufMu: sync.Mutex{},
		closed:    false,
		closeChan: make(chan struct{}, 10),
	}
}
func (c *ServerConnection) Handle() {
	go c.handleRead()
	go c.handleClose()
}
func (c *ServerConnection) Send(articleContent string) {
	packets, err := article.Decode(articleContent)
	if err != nil {
		panic(err)
	}
	buf := bytes.Buffer{}
	for _, packet := range packets {
		_, err = buf.Write(packet.Data)
		if err != nil {
			panic(err)
		}
	}
	_, err = c.conn.Write(buf.Bytes())
	if err != nil {
		log.Println("cannot send to dest: " + err.Error())
		c.closeChan <- struct{}{}
	}
	//log.Println("sent to dest: " + string(buf.Bytes()))
}
func (c *ServerConnection) handleRead() {
	for {
		if c.closed {
			break
		}
		buf := make([]byte, 1024)
		n, err := c.conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Println("read conn error: " + err.Error())
			} else {
				log.Println("read EOF from conn")
			}
			c.closeChan <- struct{}{}
		}
		//log.Println("read from dest: " + string(buf[:n]))
		c.readBufMu.Lock()
		_, err = io.CopyN(bufio.NewWriter(&c.readBuf), bytes.NewReader(buf), int64(n))
		if err != nil {
			panic(err)
		}
		c.readBufMu.Unlock()
	}
}
func (c *ServerConnection) GetBodyToReadAndReset() string {
	c.readBufMu.Lock()
	defer c.readBufMu.Unlock()
	v := c.readBuf.Bytes()
	c.readBuf.Reset()
	str, err := article.Encode([]data.Packet{{Data: v}})
	if err != nil {
		panic(err)
	}
	return str
}
func (c *ServerConnection) handleClose() {
	select {
	case <-c.closeChan:
		log.Println("now closing connection: " + c.uuid)
		c.closed = true
		c.conn.Close()
	}
}
