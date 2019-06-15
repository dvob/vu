package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
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
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
	return cmd
}

func newImageListCmd() *cobra.Command {
	//TODO implement showAll
	var showAll bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list all base images",
		Run: func(cmd *cobra.Command, args []string) {
			images, err := mgr.ImageListAll()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			for _, image := range images {
				fmt.Println(image)
			}
		},
	}
	cmd.Flags().BoolVar(&showAll, "all", false, "Show all images instead of only base images")
	return cmd
}

func newImageRemoveCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "remove image",
		Run: func(cmd *cobra.Command, args []string) {
			name = args[0]
			err := mgr.ImageRemove(name)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
	return cmd
}
