package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func newImageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image",
		Short: "manage base images",
	}
	cmd.AddCommand(
		newImageListCmd(),
		newImageGetCmd(),
		newImageRemoveCmd(),
	)
	addPoolOption(cmd.PersistentFlags(), &virStoragePool)
	return cmd
}

func newImageGetCmd() *cobra.Command {
	var (
		name string
		url  string
	)
	cmd := &cobra.Command{
		Use:   "get <name> <url>",
		Short: "create new base image form url",
		Long:  "get a base image from either an http:// or file:// url  and store it in the storage pool",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			name = args[0]
			url = args[1]

			err := mgr.CreateBaseImage(name, url)
			if err != nil {
				errExit(err)
			}
		},
	}
	return cmd
}

func newImageListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "list images",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			images, err := mgr.ImageList()
			if err != nil {
				errExit(err)
			}
			for _, image := range images {
				fmt.Println(image)
			}
		},
	}
	return cmd
}

func newImageRemoveCmd() *cobra.Command {
	var (
		name   string
		names  []string
		failed bool
	)
	cmd := &cobra.Command{
		Use:     "remove <name>...",
		Short:   "remove images",
		Aliases: []string{"rm"},
		Run: func(cmd *cobra.Command, args []string) {
			names = args

			for _, name = range names {
				err := mgr.ImageRemove(name)
				if err != nil {
					failed = true
					errPrint(err)
				}
			}
			if failed {
				os.Exit(1)
			}
		},
	}
	return cmd
}
