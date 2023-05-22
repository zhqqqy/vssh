package vssh

import (
	"context"
	"fmt"
	pb "github.com/cheggaaa/pb/v3"
	"io"
	"io/fs"
	"sync"
	"time"
)

type sftp struct {
	ctx         context.Context
	file        fs.File
	remotePath  string
	respChan    chan *Response
	respTimeout time.Duration
	action      int
}

type progressWriter struct {
	writer io.Writer
	bar    *pb.ProgressBar
	total  int64
}

func (pw *progressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.writer.Write(p)
	pw.bar.Add(n)
	return
}

func (s sftp) run(v *VSSH) {
	var wg sync.WaitGroup
	for client := range v.clients.enum() {
		client := client

		wg.Add(1)
		go func() {
			defer wg.Done()
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
		}()
	}
	wg.Wait()
	close(s.respChan)
}

func (s sftp) errResp(addr string, err error) {
	err = fmt.Errorf("client [%s]: %v", addr, err)
	s.respChan <- &Response{id: addr, err: err}
}
