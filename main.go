// ShiTTYSSH
// ..... mostly adapted from reverseSSH code.

// reverseSSH - a lightweight ssh server with a reverse connection feature
// Copyright (C) 2021  Ferdinor <ferdinor@mailbox.org>

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"io"
	"net"
	"os"
	"time"

	"github.com/gliderlabs/ssh"
	"github.com/pkg/sftp"
)

func main() {
	forwardHandler := &ssh.ForwardedTCPHandler{}

	server := ssh.Server{
		Handler:                       createSSHSessionHandler("sh"),
		LocalPortForwardingCallback:   createLocalPortForwardingCallback(),
		ReversePortForwardingCallback: createReversePortForwardingCallback(),
		SessionRequestCallback:        createSessionRequestCallback(),
		ChannelHandlers: map[string]ssh.ChannelHandler{
			"direct-tcpip": ssh.DirectTCPIPHandler,
			"session":      ssh.DefaultSessionHandler,
		},
		RequestHandlers: map[string]ssh.RequestHandler{
			"tcpip-forward":        forwardHandler.HandleSSHRequest,
			"cancel-tcpip-forward": forwardHandler.HandleSSHRequest,
		},
		SubsystemHandlers: map[string]ssh.SubsystemHandler{
			"sftp": createSFTPHandler(),
		}}
	server.Serve(&STDIOListener{})
}

type STDIOAddr struct {
	S string
}
type STDIOListener struct {
	Accepted bool
	S        string
}
type STDIOConn struct {
	S string
	F *os.File
}

//Addr
func (addr STDIOAddr) Network() string {
	return addr.S
}
func (addr STDIOAddr) String() string {
	return addr.S
}

//Listener
func (listen *STDIOListener) Accept() (net.Conn, error) {

	if listen.Accepted == false {
		b := make([]byte, 1)
		for {
			os.Stdin.Read(b)
			if b[0] == 'G' {
				listen.Accepted = true

				//Close STDOUT to the rest of the program.
				stdout := os.Stdout
				devnull, _ := os.Open("/dev/null")
				os.Stdout = devnull

				conn := &STDIOConn{F: stdout}
				return conn, nil
			}
		}
	}

	c := make(chan struct{})
	<-c

	return nil, nil
}

func createLocalPortForwardingCallback() ssh.LocalPortForwardingCallback {
	return func(ctx ssh.Context, dhost string, dport uint32) bool {
		return true
	}
}

func createReversePortForwardingCallback() ssh.ReversePortForwardingCallback {
	return func(ctx ssh.Context, host string, port uint32) bool {
		return true
	}
}

func createSessionRequestCallback() ssh.SessionRequestCallback {
	return func(sess ssh.Session, requestType string) bool {
		return true
	}
}

func createSFTPHandler() ssh.SubsystemHandler {
	return func(s ssh.Session) {
		server, err := sftp.NewServer(s)
		if err != nil {
			return
		}

		if err := server.Serve(); err == io.EOF {
			server.Close()
		} else if err != nil {
		}
	}
}

func (listen STDIOListener) Close() error {
	return nil
}
func (listen STDIOListener) Addr() net.Addr {
	return &STDIOAddr{S: "STDIOPipe"}
}

func (conn STDIOConn) Read(b []byte) (n int, err error) {
	return os.Stdin.Read(b)
}

func (conn STDIOConn) Write(b []byte) (n int, err error) {
	return conn.F.Write(b)
}

func (conn STDIOConn) Close() error {
	if err := os.Stdin.Close(); err != nil {
		return err
	}
	os.Exit(0)
	return conn.F.Close()
}

func (conn STDIOConn) LocalAddr() net.Addr {
	return &STDIOAddr{S: "STDIOPipeLocal"}
}

func (conn STDIOConn) RemoteAddr() net.Addr {
	return &STDIOAddr{S: "STDIOPipeRemote"}
}

func (conn STDIOConn) SetDeadline(t time.Time) error {
	return nil
}

func (conn STDIOConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (conn STDIOConn) SetWriteDeadline(t time.Time) error {
	return nil
}
