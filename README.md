# vu
vu (**v**irtual machin **up**) is a CLI tool to quickly spin up a virtual machines.


```bash
# get image
vu image add https://cloud-images.ubuntu.com/daily/server/bionic/current/bionic-server-cloudimg-amd64.img

# list images
vu image ls

# create new VM
vu create bionic-server-cloudimg-amd64.img mytest1

# list VMs
vu list

# connect to vm
ssh mytest1

# remove VM
vu rm mytest1
```
