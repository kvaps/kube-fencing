FROM alpine

RUN wget https://storage.googleapis.com/kubernetes-release/release/$(wget -O- https://dl.k8s.io/release/stable.txt 2>/dev/null)/bin/linux/amd64/kubectl -O /usr/local/bin/kubectl \
 && chmod +x /usr/local/bin/kubectl

ADD fencing-switcher.sh /usr/local/bin/fencing-switcher.sh

CMD ["/usr/local/bin/fencing-switcher.sh"]
