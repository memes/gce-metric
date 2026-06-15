FROM alpine:3.24.0 as ca
RUN apk --no-cache add ca-certificates-bundle=20260413-r0

FROM scratch
COPY --from=ca /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY gce-metric /gce-metric
ENTRYPOINT ["/gce-metric"]
