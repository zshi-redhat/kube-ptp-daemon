FROM fedora:30
ADD . /usr/src/kube-ptp-daemon

WORKDIR /usr/src/kube-ptp-daemon

RUN yum install -y ethtool make hwdata golang
RUN make clean && make

WORKDIR /

CMD ["/usr/src/kube-ptp-daemon/bin/ptp"]
