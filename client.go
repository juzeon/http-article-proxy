package main

import (
	"bytes"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"http-article-proxy/article"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Client struct {
	port       int
	url        string
	httpClient *resty.Client
}
type ClientConnection struct {
	conn        net.Conn
	uuid        string
	url         string
	httpClient  *resty.Client
	sendBuf     bytes.Buffer
	sendBufMu   sync.Mutex
	receiveChan chan []byte
	closed      bool
	closeChan   chan struct{}
	closeMu     sync.Mutex
}

func NewClient(port int, url string) *Client {
	return &Client{
		port:       port,
		url:        url,
		httpClient: resty.New().SetTimeout(10 * time.Second),
	}
}
func (c *Client) Serve() {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(c.port))
	if err != nil {
		panic(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("cannot accept connection: " + err.Error())
		}
		connection := NewClientConnection(conn, c.url, c.httpClient)
		go connection.Handle()
	}
}

func NewClientConnection(conn net.Conn, url string, httpClient *resty.Client) *ClientConnection {
	if strings.HasSuffix(url, "/") {
		url = url[:len(url)-1]
	}
	u := uuid.New().String()
	log.Println("new connection with uuid: " + u)
	return &ClientConnection{
		conn:        conn,
		uuid:        u,
		url:         url,
		httpClient:  httpClient,
		sendBuf:     bytes.Buffer{},
		sendBufMu:   sync.Mutex{},
		receiveChan: make(chan []byte, 512),
		closed:      false,
		closeChan:   make(chan struct{}, 10),
		closeMu:     sync.Mutex{},
	}
}
func (c *ClientConnection) Handle() {
	go c.handleRead()
	go c.handleTransfer()
	go c.handleWrite()
	go c.handleClose()
}
func (c *ClientConnection) handleRead() {
	for {
		if c.closed {
			break
		}
		buf := make([]byte, 1024)
		n, err := c.conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				log.Println("read conn error: " + err.Error())
			}
			c.closeChan <- struct{}{}
		}
		//log.Println("read from inbound: " + string(buf[:n]))
		c.sendBufMu.Lock()
		_, err = c.sendBuf.Write(buf[:n])
		if err != nil {
			panic(err)
		}
		c.sendBufMu.Unlock()
	}
}
func (c *ClientConnection) handleWrite() {
	for dat := range c.receiveChan {
		_, err := c.conn.Write(dat)
		if err != nil {
			log.Println("write conn error: " + err.Error())
			c.closeChan <- struct{}{}
		}
		//log.Println("wrote to inbound: " + string(dat))
	}
}
func (c *ClientConnection) handleTransfer() {
	for {
		if c.closed {
			break
		}
		c.sendBufMu.Lock()
		bytesToSend := c.sendBuf.Bytes()
		c.sendBuf.Reset()
		c.sendBufMu.Unlock()
		articleToSend := ""
		if len(bytesToSend) > 0 {
			//log.Println("bytes to send: " + string(bytesToSend))
			result, err := article.Encode(bytesToSend)
			if err != nil {
				panic(err)
			}
			articleToSend = result
		}
		//log.Println("article to send: " + articleToSend)
		resp, err := c.httpClient.R().SetBody(articleToSend).Post(c.url + "/" + c.uuid)
		if err != nil {
			log.Println("cannot send http request: " + err.Error())
			c.closeChan <- struct{}{}
			break
		}
		if resp.Header().Get("X-Connection") == "closed" {
			log.Println("X-Connection == closed")
			c.closeChan <- struct{}{}
			break
		}
		articleReceived := resp.String()
		//log.Println("article received: " + articleReceived)
		if articleReceived == "" {
			continue
		}
		bytesReceived, err := article.Decode(articleReceived)
		if err != nil {
			panic(err)
		}
		c.closeMu.Lock()
		if !c.closed {
			c.receiveChan <- bytesReceived
		}
		c.closeMu.Unlock()
	}
}
func (c *ClientConnection) handleClose() {
	select {
	case <-c.closeChan:
		c.closeMu.Lock()
		log.Println("now closing connection: " + c.uuid)
		c.closed = true
		c.conn.Close()
		close(c.receiveChan)
		c.closeMu.Unlock()
	}
}
