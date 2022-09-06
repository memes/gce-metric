# GCP custom metric generator

[![Go Reference](https://pkg.go.dev/badge/github.com/memes/gce-metric.svg)](https://pkg.go.dev/github.com/memes/gce-metric)
[![Go Report Card](https://goreportcard.com/badge/github.com/memes/gce-metric)](https://goreportcard.com/report/github.com/memes/gce-metric)

A synthetic metric generator for Google Cloud that will create a time-series of
artificial metrics that can be consumed by an autoscaler, or other metric-bound
resource. By default, the application will automatically try to generate a time-series
with metric labels that match expectations for [gce_instance], or [gke_container]
metrics if an appropriate Google Cloud environment is detected, with fallback to
[generic_node] metrics.

## Usage

The application has three-forms of operation; *generator*, *list*, and *delete*.

### Generator

```sh
gce-metric waveform [FLAGS] NAME
```

- *waveform* is one of sawtooth, sine, square, or triangle, and sets the pattern for the metrics (see images below)
- **NAME** is the custom metric type to add to GCP; this name must not conflict with existing metrics provided by GCP, and convention suggests that it be of the form `custom.googleapis.com/name` - see GCP [creating metrics] docs for details

All options have a default value, but can be changed through flags:

- `--floor N` sets the minimum value for the cycles, can be an integer or floating point value
- `--ceiling N` sets the maximum value for the cycles, can be an integer of floating point value
- `--period T` sets the duration for one complete cycle from floor to ceiling, must be valid Go duration string (see [time.Duration])
- `--sample T` sets the interval between sending metrics to Google Monitoring, must be valid Go duration string (see [time.ParseDuration])
- `--verbose` set the logging levels to include more details
- `--integer` forces the generated metrics to be integers, making them 'steppier'

> **NOTE:** Custom metric names can be reused as long as the type of the metric
> doesn't change; i.e. if you created a metric with floating point values, and
> then try to use `--integer` with the same metric name it will fail.

If the application is executed on a GCE VM, or in a container on a GCE VM, the
project identifier (and other details) will be pulled from metadata. If running on a non-GCP system, you will need to ensure you are authenticated to GCP and authorised to create metric time-series.

- `--project ID` will set (or override discovered) project ID for the metrics
- `--metric-labels key1=value1,key2=value2` and `--resource-labels key1=value1,key2=value2`
can be used to populate the metric and resource labels assigned to the time series, respectively.

#### Example

To generate synthetic metrics called `custom.googleapis.com/custom_metric` that start at 0 and rise to 10 in a sawtooth pattern each hour

```sh
gce-metric sawtooth --floor 0 --ceiling 10 --period 1h custom.googleapis.com/custom_metric
```

#### Sawtooth in Metrics Explorer

![Sawtooth metric in Metrics Explorer](images/sawtooth.png)

```sh
gce-metric sawtooth --floor 1 --ceiling 10 --period 20m --sample 30s custom.googleapis.com/gce_metric
```

#### Sine in Metrics Explorer

![Sine metric in Metrics Explorer](images/sine.png)

```sh
gce-metric sine --floor 1 --ceiling 10 --period 20m --sample 30s custom.googleapis.com/gce_metric
```

#### Square in Metrics Explorer

![Square metric in Metrics Explorer](images/square.png)

```sh
gce-metric square --floor 1 --ceiling 10 --period 20m --sample 30s custom.googleapis.com/gce_metric
```

#### Triangle in Metrics Explorer

![Triangle metric in Metrics Explorer](images/triangle.png)

```sh
gce-metric triangle --floor 1 --ceiling 10 --period 20m --sample 30s custom.googleapis.com/gce_metric
```

### List

To list custom metrics

```sh
gce-metric list [--verbose] [--project ID --filter FILTER]
```

- `--filter` applies a [metric filter](https://cloud.google.com/monitoring/api/v3/filters#metric-descriptor-filter) to the list. The default filter will limit the results to metrics matching `custom.googleapis.com/*` in the detected or specified project.

### Delete

To delete one or more custom metrics use

```sh
gce-metric delete [--verbose] [--project ID] NAME...
```

or combine with [list](#list) to delete all custom metrics

```sh
gce-metric list [--project ID] | xargs gce-metric delete [--project ID]
```

## Binaries

Binaries are published on the [Releases] page for Linux, macOS, and Windows. If you have Go installed locally, `go install github.com/memes/gce-metric/cmd/gce-metric` will download and install to *$GOBIN*.

A container image is also published to Docker Hub and GitHub Container Registries
that can be used in place of the binary; just append the arguments to the
`docker run` or `podman run` command.

E.g.

```sh
podman run -d --rm --name gce-metric ghcr.io/memes/gce-metric:1.2.0 sawtooth -period 1h -sample 2m
```

## Verifying releases

For each tagged release, an tarball of the source and a [syft] SBOM is created,
along with SHA256 checksums for all files. [cosign] is used to automatically generate
a signing certificate for download and verification of container images.

### Verify release files

1. Download the checksum, signature, and signing certificate file from GitHub

   ```shell
   curl -sLO https://github.com/memes/gce-metric/releases/download/1.2.0/gce-metric_1.2.0_SHA256SUMS
   curl -sLO https://github.com/memes/gce-metric/releases/download/1.2.0/gce-metric_1.2.0_SHA256SUMS.sig
   curl -sLO https://github.com/memes/gce-metric/releases/download/1.2.0/gce-metric_1.2.0_SHA256SUMS.pem
   ```

2. Verify the SHA256SUMS have been signed with [cosign]

   ```shell
   cosign verify-blob --cert gce-metric_1.2.0_SHA256SUMS.pem --signature gce-metric_1.2.0_SHA256SUMS.sig gce-metric_1.2.0_SHA256SUMS
   ```

   ```text
   verified OK
   ```

3. Download and verify files

   Now that the checksum file has been verified, any other file can be verified using `sha256sum`.

   For example

   ```shell
   curl -sLO https://github.com/memes/gce-metric/releases/download/1.2.0/gce-metric-1.2.0.tar.gz.sbom
   curl -sLO https://github.com/memes/gce-metric/releases/download/1.2.0/gce-metric_1.2.0_linux_amd64
   sha256sum --ignore-missing -c gce-metric_1.2.0_SHA256SUMS
   ```

   ```text
   gce-metric-1.2.0.tar.gz.sbom: OK
   gce-metric_1.2.0_linux_amd64: OK
   ```

### Verify container image

Use [cosign]s experimental OCI signature support to validate the container.

```shell
COSIGN_EXPERIMENTAL=1 cosign verify ghcr.io/memes/gce-metric:1.2.0
```

[gce_instance]: https://cloud.google.com/monitoring/api/resources#tag_gce_instance
[gke_container]: https://cloud.google.com/monitoring/api/resources#tag_gke_container
[generic_node]: https://cloud.google.com/monitoring/api/resources#tag_generic_node
[creating metrics]: https://cloud.google.com/monitoring/custom-metrics/creating-metrics#custom_metric_names
[time.ParseDuration]: https://golang.org/pkg/time/#ParseDuration
[Releases]: https://github.com/memes/gce-metric/releases
[cosign]: https://github.com/SigStore/cosign
[syft]: https://github.com/anchore/syft
