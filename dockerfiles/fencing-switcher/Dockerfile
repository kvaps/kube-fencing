FROM alpine

RUN apk add --no-cache curl
ADD fencing-switcher.sh /usr/local/bin/fencing-switcher.sh

CMD ["/usr/local/bin/fencing-switcher.sh"]
