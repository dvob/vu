# vu
`vu` (**v**irtual machin **u**p) is a CLI tool which lets you spin up virtual machines quickly. Its goal is to spin up a new VM and connect to it via SSH without going through a cumbersome install wizard and without maintaining special VM images which support automatic installation ourselves.
To achieve this `vu` relies on VM images which contain [cloud-init](https://cloudinit.readthedocs.io/en/latest/). Cloud-init runs during the first boot, reads the configuration from a datasource and then configures the VM accordingly (user, network, etc.).
The datasource is usually provided by a cloud provider. To use the cloud-init images locally we make use of the [NoCloud](https://cloudinit.readthedocs.io/en/latest/topics/datasources/nocloud.html) datasource where we provide the configuration as an attached CDROM.

`vu` usually does the following steps to start a VM:
* Create a cloud-init config (user data, metadata, network configuration) based on the local user (username, ssh key)
* Create an ISO image with the cloud-init config in it
* Clone an image for the new VM from a base image (see https://libvirt.org/kbase/backing_chains.html)
* Start a VM with cloned image and the the ISO image attached as CDROM

## Quick start
* Install libvirtd
```
# on Ubuntu
sudo apt install libvirt-daemon-system
sudo adduser $USER libvirt
```

* Run a VM
```bash
# get  a base image
vu image add https://cloud-images.ubuntu.com/minimal/daily/focal/current/focal-minimal-cloudimg-amd64.img

# create and start a new VM
vu create focal-minimal-cloudimg-amd64.img mytest1

# list the VMs
vu list
```

To connect to the VM it is recommended to install the [Libvirt NSS module](https://libvirt.org/nss.html) and then you can do:
```
ssh mytest1
```

If you don't want to use the Libvirt NSS module or you specified a static IP on create with `--ip` then you have to use the IP to connect to the VM.

## Images
To find base images you can search for `cloud init images` and then look out for images in the `qcow2` format. Usually they have the `.img` or `.qcow2` file ending. The following link provides a good overview on where you can find cloud-init images: https://docs.openstack.org/image-guide/obtain-images.html

If you have found an appropriate image you can download it (add it to the base images) with `vu image add`. I tested `vu` with the following images:
```shell
# add ubuntu image
vu image add https://cloud-images.ubuntu.com/daily/server/bionic/current/bionic-server-cloudimg-amd64.img
vu image add https://cloud-images.ubuntu.com/minimal/daily/focal/current/focal-minimal-cloudimg-amd64.img

# rocky linux image
vu image add https://download.rockylinux.org/pub/rocky/8.5/images/Rocky-8-GenericCloud-8.5-20211114.2.x86_64.qcow2

# add centos image
vu image add https://cloud.centos.org/centos/8/x86_64/images/CentOS-8-GenericCloud-8.3.2011-20201204.2.x86_64.qcow2

# add debian image
vu image add http://cdimage.debian.org/cdimage/openstack/current/debian-10-openstack-amd64.qcow2
```

### Storage location
`vu` uses three [storage pools](https://libvirt.org/storage.html) to store the images:
* `base` for base images
* `vm` for vm instances
* `config` for config ISOs (cidata for cloudinit)

If these storage pools do not yet exist `vu` creates them on the fly as [directory pool](https://libvirt.org/storage.html#StorageBackendDir) under `/var/lib/libvirt/images/vu/{base,config,vm}`.

## Shell completion
```
source <( vu completion bash )
```
