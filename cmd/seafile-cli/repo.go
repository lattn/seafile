package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	cli "github.com/urfave/cli/v2"

	"github.com/lattn/seafile"
)

var listRepoCmd = &cli.Command{
	Name:  "list-repo",
	Usage: "list repos",
	Action: func(cctx *cli.Context) error {
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

		tw := tabwriter.NewWriter(os.Stdout, 2, 2, 2, ' ', 0)
		_, _ = fmt.Fprintf(tw, "ID\tPermision\tMTime\tName\n")
		for _, repo := range repos {
			_, _ = fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", repo.ID, repo.Permission, time.Unix(int64(repo.MTime), 0).Format("2006-01-02 15:04:05"), repo.Name)
		}
		_ = tw.Flush()

		return nil
	},
}
