package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/lxbot/lxlib/v2/lxtypes"
)

type Process struct {
	cmd    *exec.Cmd
	stdIn  *io.WriteCloser
	stdOut *io.ReadCloser

	MessageCh *chan *InternalMessage
}

type PartialMessage struct {
	ID    string            `json:"id"`
	Event lxtypes.EventType `json:"event"`
}

type InternalMessage struct {
	ID        string
	Origin    string
	EventType lxtypes.EventType
	Body      string
}

const (
	ExitEvent lxtypes.EventType = "exit"
)

func mustNewProcess(filePath string) *Process {
	cmd := exec.Command(filePath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalln("could not open stdin")
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalln("could not open stdout")
	}
	cmd.Stderr = os.Stderr

	ch := make(chan *InternalMessage)
	p := &Process{
		cmd:       cmd,
		stdIn:     &stdin,
		stdOut:    &stdout,
		MessageCh: &ch,
	}
	go p.listen()
	go p.run()

	_ = p.waitForReady()
	// TODO: stdio以外の通信タイプに対応する場合はここで判定する

	return p
}

func newDumbProcess() *Process {
	ch := make(chan *InternalMessage)
	p := &Process{
		cmd:       nil,
		stdIn:     nil,
		stdOut:    nil,
		MessageCh: &ch,
	}
	return p
}

func (this *Process) Dispose() {
	if this.cmd == nil {
		return
	}
	_ = syscall.Kill(-this.cmd.Process.Pid, syscall.SIGTERM)
}

func (this *Process) Origin() string {
	if this.cmd == nil {
		return "dumb"
	}
	return this.cmd.Path
}

func (this *Process) run() {
	if err := this.cmd.Run(); err != nil {
		log.Fatalln("start error:", this.Origin(), err)
	}
	*this.MessageCh <- &InternalMessage{
		ID:        "",
		EventType: ExitEvent,
		Body:      "",
		Origin:    this.Origin(),
	}
}

func (this *Process) listen() {
	if this.stdOut == nil {
		return
	}

	s := bufio.NewScanner(*this.stdOut)
	for {
		for s.Scan() {
			line := s.Text()
			if strings.HasPrefix(line, "{") && strings.HasSuffix(line, "}") {
				this.onMessage(line)
			}
		}
		if s.Err() != nil {
			log.Println("error:", s.Err())
		}
	}
}

func (this *Process) onMessage(body string) {
	msg := new(PartialMessage)
	if err := json.Unmarshal([]byte(body), msg); err != nil {
		log.Println("error:", err)
	}

	*this.MessageCh <- &InternalMessage{
		ID:        msg.ID,
		EventType: msg.Event,
		Body:      body,
		Origin:    this.Origin(),
	}
}

func (this *Process) waitForReady() string {
	for {
		msg := <-*this.MessageCh
		if msg.EventType == lxtypes.ReadyEvent {
			return msg.Body
		}
	}
}

func (this *Process) Write(body string) {
	if this.stdIn == nil {
		return
	}
	fmt.Fprintf(*this.stdIn, "%s\n", body)
}
