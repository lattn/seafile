package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	cli "github.com/urfave/cli/v2"

	"github.com/caeret/seafile"
)

var getFileDetail = &cli.Command{
	Name:  "file-info",
	Usage: "get file detail",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "repo",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "path",
			Required: true,
		},
	},
	Action: func(cctx *cli.Context) error {
		cfg, err := LoadConfig()
		if err != nil {
			return err
		}

		ctx := context.TODO()

		c := seafile.New(cfg.Endpoint, cfg.Token)

		repo, err := c.LibraryByName(ctx, cctx.String("repo"))
		if err != nil {
			return err
		}
		if repo == nil {
			return errors.New("repo not found")
		}

		fd, err := c.FileDetail(ctx, repo.ID, cctx.String("path"))
		if err != nil {
			return err
		}

		if fd == nil {
			fmt.Println("file not found")
			return nil
		}

		tw := tabwriter.NewWriter(os.Stdout, 2, 2, 2, ' ', 0)
		for k, v := range struct2map(fd) {
			_, _ = fmt.Fprintf(tw, "%s:\t%v\n", k, v)
		}
		_ = tw.Flush()

		return nil
	},
}
