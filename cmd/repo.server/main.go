package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type LsInput struct {
	Type         string `json:"type" jsonschema:"type of the action, one of: dirs, files, all"`
	Pattern      string `json:"pattern" jsonschema:"file system pattern of the content to include"`
	NotRecursive bool   `json:"notRecursive" jsonschema:"if true then sub directories are not included"`
}

type LsOutputItem struct {
	Path  string `json:"path" jsonschema:"path of the item"`
	IsDir bool   `json:"isDir" jsonschema:"is true if the item is a directory"`
}

type LsOutput struct {
	Items []LsOutputItem `json:"items" jsonschema:"list items of the directory"`
}

func NewLsOutput() LsOutput {
	return LsOutput{
		Items: make([]LsOutputItem, 0),
	}
}

func addLsTool(server *mcp.Server, startDir string) {
	doLs := func(ctx context.Context, req *mcp.CallToolRequest, input LsInput) (
		*mcp.CallToolResult,
		LsOutput,
		error,
	) {
		var ret LsOutput
		var err error
		switch input.Type {
		case "dirs":
			ret, err = lsGetDirs(startDir, input)
		case "files":
			ret, err = lsGetFiles(startDir, input)
		case "all":
			ret, err = lsGetAll(startDir, input)
		default:
			ret, err = LsOutput{}, fmt.Errorf("Unknown input type for ls: %s", input.Type)
		}
		return nil, ret, err
	}
	mcp.AddTool(server, &mcp.Tool{Name: "repo.ls", Description: "lists content of the choosen repository"}, doLs)
}

func getLsOutput(startDir string, matches func(name string) bool, shouldIncludeType func(bool) bool) (LsOutput, error) {
	out := NewLsOutput()
	entries, err := os.ReadDir(startDir)
	if err != nil {
		return LsOutput{}, err
	}
	for _, e := range entries {
		if !matches(e.Name()) || !shouldIncludeType(e.IsDir()) {
			continue
		}
		out.Items = append(out.Items, LsOutputItem{
			Path:  filepath.Join(startDir, e.Name()),
			IsDir: e.IsDir(),
		})
	}
	return out, nil
}

func getLsOutputRecursive(startDir string, matches func(name string) bool, shouldIncludeType func(bool) bool) (LsOutput, error) {
	var out LsOutput
	err := filepath.WalkDir(startDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == startDir {
			return nil
		}
		if !matches(d.Name()) || !shouldIncludeType(d.IsDir()) {
			return nil
		}
		out.Items = append(out.Items, LsOutputItem{Path: path, IsDir: d.IsDir()})
		return nil
	})

	if err != nil {
		return LsOutput{}, err
	}
	return out, nil
}

func getLsOutputBase(notRecursive bool, startDir, pattern string, shouldIncludeType func(bool) bool) (LsOutput, error) {
	matches := func(name string) bool {
		if pattern == "" {
			return true
		}
		ok, _ := filepath.Match(pattern, name)
		return ok
	}
	if notRecursive {
		return getLsOutput(startDir, matches, shouldIncludeType)
	} else {
		return getLsOutputRecursive(startDir, matches, shouldIncludeType)
	}
}

func lsGetDirs(startDir string, input LsInput) (LsOutput, error) {
	shouldIncludeType := func(isDir bool) bool { return isDir }
	return getLsOutputBase(input.NotRecursive, startDir, input.Pattern, shouldIncludeType)
}

func lsGetFiles(startDir string, input LsInput) (LsOutput, error) {
	shouldIncludeType := func(isDir bool) bool { return !isDir }
	return getLsOutputBase(input.NotRecursive, startDir, input.Pattern, shouldIncludeType)
}

func lsGetAll(startDir string, input LsInput) (LsOutput, error) {
	shouldIncludeType := func(isDir bool) bool { return true }
	return getLsOutputBase(input.NotRecursive, startDir, input.Pattern, shouldIncludeType)
}

func initLogger(logFile *string) {
	if logFile == nil || *logFile == "" {
		lf := "repo_server.log"
		logFile = &lf
	}
	file, err := os.OpenFile(*logFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	logger := slog.New(slog.NewTextHandler(file, nil))
	slog.SetDefault(logger)
}

func main() {
	startDir := flag.String("dir", "", "defines the start dir where the server should look into")
	logFile := flag.String("log", "", "log file to write, default is ./repo_server.log")
	flag.Parse()
	initLogger(logFile)

	// check that this server doesn't run as root!! :D
	if os.Geteuid() == 0 {
		panic("Running as root is not allowed!!!")
	}
	if startDir == nil || *startDir == "" {
		workingDir, err := os.Getwd()
		if err != nil {
			slog.Error("error while reading current working dir", "error", err)
			return
		}
		startDir = &workingDir
	}
	slog.Info("started MCP repo.server", "initDir", *startDir)
	server := mcp.NewServer(&mcp.Implementation{Name: "repo.server", Version: "v0.1.0"}, nil)
	addLsTool(server, *startDir)
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		slog.Error("error while run server", "error", err)
	}
}
