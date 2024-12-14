FROM gcr.io/distroless/static:nonroot
ARG BINARY

WORKDIR /
COPY ${BINARY} /main

USER 65532:65532

ENTRYPOINT ["/main"]