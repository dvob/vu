FROM debian:10

RUN apt-get update && \
    apt-get install -y \ 
       libvirt-daemon \
       libvirt-daemon-system \
       libvirt-clients \
       qemu \
       qemu-utils \
       qemu-kvm \
       iptables \
       dnsmasq \
       ebtables \
       dmidecode

COPY libvirtd.conf /etc/libvirt/libvirtd.conf

CMD ["/bin/bash", "-c", "virtlogd & libvirtd --listen"]
