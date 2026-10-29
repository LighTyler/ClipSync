package network

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type Peer struct {
	listenAddr string
	peerAddrs  []string

	mu       sync.Mutex
	conns    map[string]net.Conn
	msChn    chan Message
	lastHash string
}

func NewPeer(listenAddr string, peerAddrs []string) *Peer {
	return &Peer{
		listenAddr: listenAddr,
		peerAddrs:  peerAddrs,
		conns:      make(map[string]net.Conn),
		msChn:      make(chan Message, 3),
	}
}

func (p *Peer) MessageChan() <-chan Message {
	return p.msChn
}
func (p *Peer) SendMessage(msg Message) error {
	hash := sha256.Sum256(msg.Payload)
	hashStr := hex.EncodeToString(hash[:])

	p.mu.Lock()
	if hashStr == p.lastHash {
		p.mu.Unlock()
		return nil
	}

	p.lastHash = hashStr
	connsCopy := make(map[string]net.Conn, len(p.conns))
	for addr, conn := range p.conns {
		connsCopy[addr] = conn
	}
	p.mu.Unlock()

	if len(connsCopy) == 0 {
		return fmt.Errorf("network - peer - SendMessage - No connections available.")
	}

	for addr, conn := range connsCopy {
		if err := WriteMessage(conn, msg); err != nil {
			fmt.Errorf("network - peer - SendMessage - Error sending message: %v", err)
		} else {
			log.Printf("network - peer - SendMessage - Message sent to %s.", addr)
		}
	}
	return nil
}

func (p *Peer) Run(ctx context.Context) {
	go p.acceptLoop(ctx)
	for _, addr := range p.peerAddrs {
		go p.connectLoop(ctx, addr)
	}
	<-ctx.Done()
}

func (p *Peer) acceptLoop(ctx context.Context) {
	ln, err := net.Listen("tcp", p.listenAddr)
	if err != nil {
		log.Fatalf("network - peer - acceptLoop - listen error: %v", err)
	}
	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	log.Printf("Listening on %s", p.listenAddr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				continue
			}
		}
		go p.handleIncoming(conn)
	}
}

func (p *Peer) handleIncoming(conn net.Conn) {
	addr := conn.RemoteAddr().String()
	addrIP, _, _ := strings.Cut(addr, ":")

	if resolved := p.ipResolved(addrIP); !resolved {
		log.Printf("Unresolved access attempt from: %s. Permission denied.", addr)
		conn.Close()
		return
	}

	p.mu.Lock()

	if _, exists := p.conns[addrIP]; exists {
		p.mu.Unlock()
		conn.Close()
		return
	}
	p.conns[addrIP] = conn

	p.mu.Unlock()

	log.Printf("Incoming connection from %s", addr)
	p.handleConnection(conn, addrIP)
}

func (p *Peer) connectLoop(ctx context.Context, addr string) {
	addrIP, _, _ := strings.Cut(addr, ":")
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		p.mu.Lock()
		_, hasConn := p.conns[addrIP]
		p.mu.Unlock()

		if hasConn {
			time.Sleep(2 * time.Second)
			continue
		}

		log.Printf("Connecting to %s...", addr)
		conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
		if err != nil {
			log.Printf("Failed to connect to %s: %v, retrying after 5s...", addr, err)
			time.Sleep(5 * time.Second)
			continue
		}

		p.mu.Lock()
		if _, exists := p.conns[addrIP]; exists {
			p.mu.Unlock()
			conn.Close()
			continue
		}
		p.conns[addrIP] = conn
		p.mu.Unlock()

		log.Printf("Connected to %s", addr)
		p.handleConnection(conn, addrIP)
	}
}

func (p *Peer) handleConnection(conn net.Conn, addr string) {
	defer func() {
		conn.Close()
		p.mu.Lock()
		if p.conns[addr] == conn {
			delete(p.conns, addr)
		}
		p.mu.Unlock()
		log.Printf("Connection closed from %s", addr)
	}()

	for {
		msg, err := ReadMessage(conn)
		if err != nil {
			log.Printf("Read error from %s: %v", addr, err)
			return
		}
		log.Printf("Message received from %s: %s", addr, msg.MIME)

		hash := sha256.Sum256(msg.Payload)
		hashStr := hex.EncodeToString(hash[:])

		p.mu.Lock()
		if hashStr == p.lastHash {
			p.mu.Unlock()
			continue
		}
		p.mu.Unlock()

		select {
		case p.msChn <- msg:
		default:
			log.Printf("Message channel full!")
		}
	}
}

func (p *Peer) ipResolved(addrIP string) bool {
	for _, resAddr := range p.peerAddrs {
		if resAddrIP, _, _ := strings.Cut(resAddr, ":"); resAddrIP == addrIP {
			return true
		}
	}
	return false
}
