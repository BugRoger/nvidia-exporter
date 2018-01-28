FROM golang:1.9.2-alpine3.6 as builder

RUN apk add --no-cache make gcc git musl-dev
RUN go get -u github.com/golang/dep/cmd/dep

WORKDIR /go/src/github.com/bugroger/nvidia-exporter
COPY . .

#RUN dep ensure -vendor-only

ARG VERSION
RUN make all

FROM alpine 

RUN mkdir -p /etc/ld.so.conf.d
RUN echo "/usr/local/cuda/lib64" >> /etc/ld.so.conf.d/cuda.conf 
RUN echo "/usr/local/nvidia/lib64" >> /etc/ld.so.conf.d/nvidia.conf

ENV PATH $PATH:/usr/local/nvidia/bin:/usr/local/cuda/bin
ENV LD_LIBRARY_PATH $LD_LIBRARY_PATH:/usr/local/nvidia/lib64/:/usr/local/cuda/lib64/

COPY --from=builder /go/src/github.com/bugroger/nvidia-exporter/bin/linux/* /

ENTRYPOINT ["/nvidia-exporter"]
