package vssh

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type sftp struct {
	ctx         context.Context
	localPath   string
	remotePath  string
	respChan    chan *Response
	respTimeout time.Duration
	action      int
}

func (s sftp) run(v *VSSH) {
	fmt.Println("made it here")
	var wg sync.WaitGroup
	for client := range v.clients.enum() {
		client := client

		wg.Add(1)

		if v.mode {
			client.connect()
		}

		if client.getErr() != nil {
			client.RLock()
			s.respChan <- &Response{id: client.addr, err: client.err}
			client.RUnlock()
			return
		}

		client.transfer(s)

		if v.mode {
			client.close()
		}
	}
}

func (s sftp) errResp(addr string, err error) {
	err = fmt.Errorf("client [%s]: %v", addr, err)
	s.respChan <- &Response{id: addr, err: err}
}
