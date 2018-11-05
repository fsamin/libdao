package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := new(cli.App)
	app.Name = "libdaogen"
	app.Usage = "generate your dao"
	app.Action = func(c *cli.Context) error {
		args := c.Args()
		for _, a := range args {
			if err := do(a, os.Stdout); err != nil {
				return fmt.Errorf("unable to process %s: %v", a, err)
			}
		}
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
