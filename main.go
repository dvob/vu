package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/dsbrng25b/cis/internal/virt"
	"github.com/dsbrng25b/cis/internal/cloud-init"
)

var version = "n/a"
var gitCommit = "n/a"
var buildTime = "n/a"

var virStoragePool string
var virNetwork string

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use: "cis",
	}
	rootCmd.AddCommand(
		newGetCmd(),
		newCreateCmd(),
		newRemoveCmd(),
		newCloudInitCmd(),
		newVersionCmd(),
	)
	rootCmd.Flags().StringVar(&virStoragePool, "pool", "default", "The storage pool to create VMs and save base images")
	rootCmd.Flags().StringVar(&virNetwork, "network", "default", "The network to attach the VM to")
	return rootCmd
}

func newGetCmd() *cobra.Command {
	var (
		name string
		url  string
	)
	cmd := &cobra.Command{
		Use:   "get <name> <url>",
		Short: "create base image form a url",
		Long:  "download a base image and store it in the storage pool",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			name = args[0]
			url = args[1]
			manager, err := virt.NewLibvirtManager(virStoragePool, virNetwork)
			if err != nil {
				fmt.Println(err)
			}

			err = manager.CreateBaseImage(name, url)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("upload successful")
			}

		},
	}
	return cmd
}

func newCreateCmd() *cobra.Command {
	var (
		name       string
		base_image string
	)
	cmd := &cobra.Command{
		Use:   "create <name> <base_image>",
		Short: "create a new VM from base image",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			name = args[0]
			base_image = args[1]

			manager, err := virt.NewLibvirtManager(virStoragePool, virNetwork)
			if err != nil {
				fmt.Println(err)
			}

			err = manager.Create(name, base_image)
			if err != nil {
				fmt.Println("error creating domain:", err)
			} else {
				fmt.Println("created domain")
			}

		},
	}
	return cmd
}

func newRemoveCmd() *cobra.Command {
	var (
		name string
	)
	cmd := &cobra.Command{
		Use:   "remove <name>",
		Short: "removes VM",
		Aliases: []string{"rm"},
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name = args[0]

			manager, err := virt.NewLibvirtManager(virStoragePool, virNetwork)
			if err != nil {
				fmt.Println(err)
			}

			err = manager.Remove(name)
			if err != nil {
				fmt.Printf("remove failed: %s\n", err)
			}

		},
	}
	return cmd
}

func newCloudInitCmd() *cobra.Command {
	var (
		name string
	)
	cmd := &cobra.Command{
		Use:   "cloud-init <name>",
		Short: "shows cloud init configuration",
		Aliases: []string{"ci"},
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name = args[0]

			userData, err := cloudinit.GetUserData(name)
			if err != nil {
				fmt.Println(err)
				return
			}

			metaData, err := cloudinit.GetMetaData(name)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Printf("### meta-data ###\n%s\n\n", metaData)
			fmt.Printf("### user-data ###\n%s\n\n", userData)

		},
	}
	return cmd
}

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "show version and build information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("version:", version)
			fmt.Println("commit:", gitCommit)
			fmt.Println("build time:", buildTime)
		},
	}
	return cmd
}

func main() {
	newRootCmd().Execute()
}
