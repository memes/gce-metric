FROM alpine:3.21.3 as ca
RUN apk --no-cache add ca-certificates-bundle=20241121-r1

FROM scratch
COPY --from=ca /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY gce-metric /gce-metric
ENTRYPOINT ["/gce-metric"]
