package main

import (
	"fmt"
	"os"

	"crawl-url/internal/crawl"
)

func main() {
	opts, err := parseCLI(os.Args[1:])
	if err == errHelp {
		printUsage()
		os.Exit(0)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n\n", err)
		printUsage()
		os.Exit(2)
	}

	if err := crawl.ValidateWorkers(opts.cfg.Workers); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if opts.cfg.MaxDepth != nil {
		if err := crawl.ValidateDepth(*opts.cfg.MaxDepth); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if err := crawl.Crawl(opts.seed, opts.cfg); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
