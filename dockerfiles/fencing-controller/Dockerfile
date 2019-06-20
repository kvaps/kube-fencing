FROM alpine

RUN wget https://storage.googleapis.com/kubernetes-release/release/$(wget -O- https://dl.k8s.io/release/stable.txt 2>/dev/null)/bin/linux/amd64/kubectl -O /usr/local/bin/kubectl \
      -O /usr/local/bin/kubectl \
 && chmod +x /usr/local/bin/kubectl

RUN apk --no-cache add jq

ADD fencing-controller.sh /usr/local/bin/fencing-controller.sh

CMD ["/usr/local/bin/fencing-controller.sh"]
