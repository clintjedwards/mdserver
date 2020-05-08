//go:generate go run mdserver/generate.go

package main

import "github.com/clintjedwards/mdserver/cli"

func main() {
	cli.RootCmd.Execute()
}
