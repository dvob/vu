package main

import (
	"fmt"
	"os"

	vu "github.com/dvob/vu/internal"
	"github.com/dvob/vu/internal/image"
	"github.com/spf13/cobra"
)

func newImageCmd(mgr *vu.Manager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image",
		Short: "manage images",
	}
	cmd.AddCommand(
		newImageListCmd(mgr),
		newImageAddCmd(mgr),
		newImageRemoveCmd(mgr),
	)
	return cmd
}

func newImageAddCmd(mgr *vu.Manager) *cobra.Command {
	var (
		name string
		url  string
	)
	cmd := &cobra.Command{
		Use:   "add URL [name]",
		Short: "Add a new image from URL",
		Long: `Adds a new image from a URL. An URL can either have a http, https or
file scheme. If no name is given the name is derived from the URL.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url = args[0]
			if len(args) > 1 {
				name = args[1]
			}
			_, err := image.AddFromURL(mgr.BaseImage, name, url, os.Stdout)
			return err
		},
	}
	return cmd
}

func newImageListCmd(mgr *vu.Manager) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "list images",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			images, err := mgr.BaseImage.List()
			if err != nil {
				return err
			}
			for _, image := range images {
				fmt.Println(image.Name)
			}
			return nil
		},
	}
	return cmd
}

func newImageRemoveCmd(mgr *vu.Manager) *cobra.Command {
	var (
		names []string
		errs  = []error{}
	)
	cmd := &cobra.Command{
		Use:     "remove <name>...",
		Short:   "remove images",
		Aliases: []string{"rm"},
		RunE: func(cmd *cobra.Command, args []string) error {
			names = args
			for _, name := range names {
				err := mgr.BaseImage.Remove(name)
				if err != nil {
					errs = append(errs, err)
				}
			}
			if len(errs) > 0 {
				return errs[0]
			}
			return nil
		},
	}
	return cmd
}
