package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/lattn/seafile"
	"github.com/urfave/cli/v2"
	"github.com/vbauerster/mpb/v8"
	"github.com/vbauerster/mpb/v8/decor"
	"golang.org/x/crypto/ssh/terminal"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

		var wg sync.WaitGroup
		p := mpb.New(mpb.WithWaitGroup(&wg), mpb.WithOutput(os.Stderr))

		for _, pair := range files {
			wg.Add(1)
			go func(pair [3]string) {
				defer wg.Done()
				err := uploadFile(ctx, p, c, repoID, pair[0], pair[1], pair[2], cctx.String("target-dir"))
				if err != nil {
					log.Printf("fail to upload file: %s", err)
				}
			}(pair)
		}

		p.Wait()

		return nil
	},
}

func uploadFile(ctx context.Context, progress *mpb.Progress, c *seafile.Client, repoID, base, dir, filename, target string) error {
	file, err := os.Open(filepath.Join(base, dir, filename))
	if err != nil {
		return err
	}
	defer file.Close()
	fi, err := file.Stat()
	if err != nil {
		return err
	}

	link, err := c.GetUploadLink(ctx, repoID)
	if err != nil {
		return err
	}

	reader := reader{r: file, ch: make(chan int)}

	bar := progress.AddBar(fi.Size(),
		mpb.PrependDecorators(
			// simple name decorator
			// decor.DSyncWidth bit enables column width synchronization
			decor.Name(filepath.Join(base, dir, filename), decor.WCSyncSpace),
			decor.CurrentKibiByte("%d", decor.WCSyncSpace),
			decor.TotalKibiByte("%d", decor.WCSyncSpace),
			decor.AverageSpeed(decor.UnitKiB, "%.1f", decor.WCSyncSpace),
		),
		mpb.AppendDecorators(
			decor.Percentage(decor.WCSyncSpace),
			decor.Elapsed(decor.ET_STYLE_GO, decor.WCSyncSpace),
		),
	)
	go func() {
		for size := range reader.ch {
			bar.IncrBy(size)
		}
	}()

	_, err = link.UploadFile(ctx, filepath.Join(target, dir, filename), reader)
	if err != nil {
		return err
	}

	return nil
}
