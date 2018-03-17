# This is a Dockerfile that builds and starts a Functron interface.
# To run Functron successfully, you have to do one more thing:
# docker run -v /var/run/docker.sock:/var/run/docker.sock
FROM golang:1.10-alpine
RUN apk update
RUN apk add docker
COPY startup.sh /root/startup.sh
COPY server.go /root/server.go
CMD sh -x /root/startup.sh
EXPOSE 8081
