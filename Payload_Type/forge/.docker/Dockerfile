FROM golang:1.23 AS builder

ARG GITHUB_TOKEN

WORKDIR /Mythic/

COPY [".", "."]

RUN make build
RUN make run_download

FROM alpine

RUN apk add make

COPY --from=builder /main /main
COPY --from=builder /collections /collections

WORKDIR /Mythic/

COPY [".", "."]

CMD make run