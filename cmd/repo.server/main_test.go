package main_test

import (
	"fmt"
	"os/exec"
	"testing"
)

func TestLsFiles(t *testing.T) {
	cmd := exec.Command("go", "run", "./main.go", "-dir", "../..", "-log", "../../temp/test.repo_server_ls.log")
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	cmd.Start()

	// Write an MCP request
	stdin.Write([]byte(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n"))

	// Read the response
	buf := make([]byte, 4096)
	n, _ := stdout.Read(buf)
	fmt.Println("Response:", string(buf[:n]))
}

func TestLsDirs(t *testing.T) {

}

func TestLsAll(t *testing.T) {

}
