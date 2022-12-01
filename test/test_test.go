package main

import (
	"fmt"
	"testing"
)

func TestGetFiles2(t *testing.T) {
	node := &FileTree{
		Label:    "root",
		Filepath: "static",
		Children: nil,
	}

	GetFiles2("/home/joseph/repo/container-dispatcher/static", "", node)

	fmt.Printf("%+v\n", node)
}
