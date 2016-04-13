// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package adb

import (
	"fmt"
	"net"
	"os/exec"
)

const (
	host = "127.0.0.1"
	port = 5037
)

func IsServerRunning() bool {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func StartServer() error {
	return exec.Command("adb", "start-server").Run()
}
