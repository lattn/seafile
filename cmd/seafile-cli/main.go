package main

import (
	"fmt"
	"os"
	"strings"

	cli "github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/lattn/seafile"
)

func main() {
	app := cli.NewApp()
	app.Name = "seafile-cli"
	app.Commands = []*cli.Command{
		initCmd,
		uploadDirCmd,
		listRepoCmd,
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var initCmd = &cli.Command{
	Name:  "generate-token",
	Usage: "generate seafile api token",
	Action: func(cctx *cli.Context) error {
		var endpoint, username, password string
		fmt.Printf("Enter the endpoint: ")
		_, _ = fmt.Scanln(&endpoint)
		fmt.Printf("Enter the username: ")
		_, _ = fmt.Scanln(&username)
		fmt.Printf("Enter the password: ")
		p, err := terminal.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return err
		}
		fmt.Println("[***]")
		password = string(p)

		endpoint = strings.TrimSpace(endpoint)
		username = strings.TrimSpace(username)
		password = strings.TrimSpace(password)

		token, err := seafile.GenerateToken(endpoint, username, password)
		if err != nil {
			return err
		}

		cfg := Config{
			Endpoint: endpoint,
			Token:    token,
		}
		err = cfg.Save()
		if err != nil {
			return err
		}

		fmt.Println("config saved.")

		return nil
	},
}
