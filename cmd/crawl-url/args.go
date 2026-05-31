package main

import "strings"

// splitSeedAndFlags separates the seed URL from flags in any order.
func splitSeedAndFlags(args []string) (seed string, flagArgs []string) {
	for _, arg := range args {
		if seed == "" && isHTTPURL(arg) {
			seed = arg
			continue
		}
		flagArgs = append(flagArgs, arg)
	}
	return seed, flagArgs
}

func isHTTPURL(arg string) bool {
	return strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://")
}
