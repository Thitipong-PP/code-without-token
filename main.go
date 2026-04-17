package main

import (
	"fmt"
	"os"

	"github.com/Thitipong-PP/code-without-token/internal/builder"
	"github.com/Thitipong-PP/code-without-token/internal/cli"
	"github.com/Thitipong-PP/code-without-token/internal/output"
)

func main() {
	cfg := cli.Parse()

	result, err := builder.Build(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	output.Print(result)
	output.CopyToClipboard(result)
}