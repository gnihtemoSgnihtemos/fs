package cmd

import "fmt"

type Test struct{ opts }

func (c *Test) Execute(args []string) error {
	cfg := mustReadConfig(c.Config)
	json, err := cfg.JSON()
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", json)
	return nil
}
