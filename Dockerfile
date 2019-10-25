FROM golang:1.13-alpine3.10 as builder

RUN apk add --no-cache make gcc git musl-dev

WORKDIR /go/src/github.com/mrpk1906/nvidia-exporter
COPY . .

ARG VERSION
RUN make all

FROM alpine

RUN mkdir -p /etc/ld.so.conf.d
RUN echo "/usr/local/cuda/lib64" >> /etc/ld.so.conf.d/cuda.conf 
RUN echo "/usr/local/nvidia/lib64" >> /etc/ld.so.conf.d/nvidia.conf

ENV PATH $PATH:/usr/local/nvidia/bin:/usr/local/cuda/bin
ENV LD_LIBRARY_PATH $LD_LIBRARY_PATH:/usr/local/nvidia/lib64/:/usr/local/cuda/lib64/

COPY --from=builder /go/src/github.com/mrpk1906/nvidia-exporter/bin/linux/* /

ENTRYPOINT ["/nvidia-exporter"]
