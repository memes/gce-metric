FROM golang:1.14.4 AS builder
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -o gce-metric

FROM alpine:3.12.0
ARG COMMIT_SHA="uncommitted"
ARG TAG_NAME="untagged"
LABEL maintainer="Matthew Emes <memes@matthewemes.com>" \
      org.opencontainers.image.title="gce-metric" \
      org.opencontainers.image.authors="memes@matthewemes.com" \
      org.opencontainers.image.description="Generate a periodic set of metrics for testing autoscaling in GCP" \
      org.opencontainers.image.url="https://github.com/memes/gce-metric" \
      org.opencontainers.image.source="https://github.com/memes/gce-metric/tree/${COMMIT_SHA}" \
      org.opencontainers.image.documentation="https://github.com/memes/gce-metric/tree/${COMMIT_SHA}/README.md" \
      org.opencontainers.image.version="${TAG_NAME}" \
      org.opencontainers.image.revision="${COMMIT_SHA}" \
      org.opencontainers.image.licenses="MIT" \
      org.label-schema.schema-version="1.0" \
      org.label-schema.schema.name="gce-metric" \
      org.label-schema.description="Generate a periodic set of metrics for testing autoscaling in GCP" \
      org.label-schema.url="https://github.com/memes/gce-metric" \
      org.label-schema.vcs-url="https://github.com/memes/gce-metric/tree/${COMMIT_SHA}" \
      org.label-schema.usage="https://github.com/memes/gce-metric/tree/${COMMIT_SHA}/README.md" \
      org.label-schema.version="${TAG_NAME}" \
      org.label-schema.vcs_ref="${COMMIT_SHA}"

RUN apk --no-cache add ca-certificates=20191127-r3
WORKDIR /run
COPY --from=builder /src/gce-metric /usr/local/bin/
ENTRYPOINT ["/usr/local/bin/gce-metric"]
