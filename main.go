package main

import (
	"fmt"
	"github.com/linuxsuren/ssh-hosts/cmd"
	"os"
)

func main() {
	array := []string{"a"}
	fmt.Println(array[1:])
	if err := cmd.NewRootCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
