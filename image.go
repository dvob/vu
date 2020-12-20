package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/digitalocean/go-libvirt"
	"github.com/dvob/vu/internal/image"
	manager "github.com/dvob/vu/internal/image/libvirt"
	"github.com/spf13/cobra"
)

type libvirtOptions struct {
	// TODO: add connect URI
	libvirt *libvirt.Libvirt
}

func (o *libvirtOptions) complete() error {
	c, err := net.DialTimeout("unix", "/var/run/libvirt/libvirt-sock", 2*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to libvirtd: %s", err)
	}

	o.libvirt = libvirt.New(c)
	if err := o.libvirt.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	return nil
}

type imageOptions struct {
	pool    string
	path    string
	libvirt libvirtOptions
	image.Manager
}

func newImageOptions() *imageOptions {
	return &imageOptions{
		pool: "vu_base_images",
		path: filepath.Join(os.Getenv("HOME"), ".local", "share", "vu", "base_image"),
	}
}

func (o *imageOptions) bindFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&o.pool, "pool", o.pool, "Storage pool for base images.")
	cmd.PersistentFlags().StringVar(&o.path, "pool-path", o.path, "Path for the storage pool for base images.")
}

func (o *imageOptions) complete() error {
	err := o.libvirt.complete()
	if err != nil {
		return err
	}

	o.Manager = manager.New(o.pool, o.path, o.libvirt.libvirt)
	return nil
}

func newImageCmd() *cobra.Command {
	o := newImageOptions()
	cmd := &cobra.Command{
		Use:   "image",
		Short: "manage images",
	}
	cmd.AddCommand(
		newImageListCmd(o),
		newImageAddCmd(o),
		newImageRemoveCmd(o),
	)
	o.bindFlags(cmd)
	return cmd
}

func newImageAddCmd(o *imageOptions) *cobra.Command {
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
			err := o.complete()
			if err != nil {
				return err
			}
			url = args[0]
			if len(args) > 1 {
				name = args[1]
			}
			_, err = image.AddFromURL(o, name, url, os.Stdout)
			return err
		},
	}
	return cmd
}

func newImageListCmd(o *imageOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "list images",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.complete()
			if err != nil {
				return err
			}
			images, err := o.List()
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

func newImageRemoveCmd(o *imageOptions) *cobra.Command {
	var (
		names []string
		errs  = []error{}
	)
	cmd := &cobra.Command{
		Use:     "remove <name>...",
		Short:   "remove images",
		Aliases: []string{"rm"},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := o.complete()
			if err != nil {
				return err
			}
			names = args
			for _, name := range names {
				err := o.Remove(name)
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
