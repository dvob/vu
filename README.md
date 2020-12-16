# vu
vu (**v**irtual machin **up**) is a CLI tool to quickly spin up a virtual machine.

```bash
# get image
vu image get https://cloud-images.ubuntu.com/daily/server/bionic/current/bionic-server-cloudimg-amd64.img ubuntu_bionic_current

# list images
vu image ls

# create new VM
vu create mytest1

# connect to vm
ssh mytest1
```
