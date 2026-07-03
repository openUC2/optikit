package main

import (
	"fmt"
	"os"

	"github.com/openUC2/optikit/internal/clients/build123d"
)

func main() {
	c, err := build123d.New()
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		panic(err)
	}
	defer func() {
		if err := c.Close(); err != nil {
			fmt.Fprint(os.Stderr, err)
			panic(err)
		}
	}()

	inputFile := "requirements.direct.txt"
	if len(os.Args) > 1 && os.Args[1] != "" {
		inputFile = os.Args[1]
	}
	result, err := c.PipFreeze(inputFile)
	if err != nil {
		fmt.Fprint(os.Stderr, err)
		panic(err)
	}
	fmt.Println(string(result))

	if len(os.Args) <= 1 || os.Args[2] == "" {
		return
	}
	const perms = 0o644
	if err = os.WriteFile(os.Args[2], result, perms); err != nil {
		fmt.Fprint(os.Stderr, err)
		panic(err)
	}
}
