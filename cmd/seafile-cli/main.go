package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/lattn/seafile"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/ssh/terminal"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	app := cli.NewApp()
	app.Name = "seafile-cli"
	app.Commands = []*cli.Command{
		initCmd,
		uploadDirCmd,
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

var uploadDirCmd = &cli.Command{
	Name:      "upload",
	Usage:     "upload files or directories",
	ArgsUsage: "[dir_to_upload | file_to_upload] ...",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "repo",
			Usage:    "target repo",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "target-dir",
			Usage:    "target directory to upload files in specified dir",
			Required: true,
		},
	},
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() {
			return cli.ShowSubcommandHelp(cctx)
		}

		cfg, err := LoadConfig()
		if err != nil {
			return err
		}

		ctx := context.TODO()

		c := seafile.New(cfg.Endpoint, cfg.Token)
		repos, err := c.ListLibraries(ctx)
		if err != nil {
			return err
		}

		repoID := ""
		for _, repo := range repos {
			if repo.Name == cctx.String("repo") {
				repoID = repo.ID
			}
		}
		if repoID == "" {
			return errors.New("repo not found")
		}

		var files [][3]string
		for _, fileOrDir := range cctx.Args().Slice() {
			fi, err := os.Stat(fileOrDir)
			if err != nil {
				return err
			}
			if fi.IsDir() {
				err = filepath.WalkDir(fileOrDir, func(path string, d fs.DirEntry, err error) error {
					if d.IsDir() {
						return nil
					}
					if strings.HasPrefix(path, ".") || strings.HasPrefix(filepath.Base(path), ".") {
						return nil
					}
					files = append(files, [3]string{fileOrDir, strings.TrimPrefix(filepath.Dir(path), filepath.Clean(fileOrDir)), filepath.Base(path)})
					return nil
				})
				if err != nil {
					return err
				}
			} else {
				files = append(files, [3]string{filepath.Dir(fileOrDir), "", filepath.Base(fileOrDir)})
			}
		}

		for _, pair := range files {
			fmt.Println(pair)
			log.Printf("upload file \"%s\" in \"%s\"", pair[2], filepath.Join(pair[0], pair[1]))

			err := uploadFile(ctx, c, repoID, pair[0], pair[1], pair[2], cctx.String("target-dir"))
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func uploadFile(ctx context.Context, c *seafile.Client, repoID, base, dir, filename, target string) error {
	file, err := os.Open(filepath.Join(base, dir, filename))
	if err != nil {
		return err
	}
	defer file.Close()

	link, err := c.GetUploadLink(ctx, repoID)
	if err != nil {
		return err
	}

	_, err = link.UploadFile(ctx, filepath.Join(target, dir, filename), file)
	if err != nil {
		return err
	}

	return nil
}
