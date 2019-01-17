# Log Aggregator (naming is hard) [![CircleCI](https://circleci.com/gh/wearefair/log-aggregator.svg?style=svg)](https://circleci.com/gh/wearefair/log-aggregator)

Reads logs from Journald, annotates/transforms, and forwards to AWS Kinesis Firehose.

## What is this for?

This is a lightweight process for reading all logs from the system, and forwarding them to a central location for
searching and long-term archival.

It also supports performing minimal transformation and annotation operations on each log entry.

### Is this designed to be generally usable outside of Fair?

While the existing code is primarily designed to fit our use-case, it is entirely interface driven, and can be used as a package to power your own processing pipeline.

Pull requests are welcome!


The main abstraction unit is a [Pipeline](https://godoc.org/github.com/wearefair/log-aggregator/pkg/pipeline#Pipeline), which is made up of the following components:

- [Source](https://godoc.org/github.com/wearefair/log-aggregator/pkg/sources#Source): produces log records
- [Destination](https://godoc.org/github.com/wearefair/log-aggregator/pkg/destinations#Destination): saves log records, and reports progress
- [Cursor](https://godoc.org/github.com/wearefair/log-aggregator/pkg/cursor#DB): provides a starting point for the `Source`, and persists the `Destination` progress so that processing can be resumed on restarts from a last-known checkpoint
- [Transformers](https://godoc.org/github.com/wearefair/log-aggregator/pkg/transform#Transformer): transform log records prior to sending to the destination.


#### Example Pipeline
Here is a simple pipeline that uses a mock source and destination, and applies the JSON transformer.

```go
package main

import (
	"time"

	"github.com/wearefair/log-aggregator/pkg/cursor"
	"github.com/wearefair/log-aggregator/pkg/destinations/stdout"
	"github.com/wearefair/log-aggregator/pkg/pipeline"
	"github.com/wearefair/log-aggregator/pkg/sources/mock"
	"github.com/wearefair/log-aggregator/pkg/transform"
	"github.com/wearefair/log-aggregator/pkg/transform/json"
)

func main() {
  cursor, _ := cursor.New("/var/log/log-aggregator.cursor")
  source = mock.New(time.Second * 2) # produce a log every 2 seconds
  destination = stdout.New()

  logPipeline, _ := pipeline.New(pipeline.Config{
    MaxBuffer:    200,
    Cursor:       cursor,
    Input:        source,
    Destination:  destination,
    Transformers: []transform.Transformer{json.Transform},
  })

  logPipeline.start()
}
```

### Features/Design

- Input: Journald
- Output: AWS Kinesis Firehose
- Transformations
  - AWS: adds `aws.instance_id`, `aws.local_hostname`, `aws.local_ipv4`
  - Journal: Rename `MESSAGE` field to `log`
  - K8s: Add Pod metadata if the log comes from a Kubernetes Pod
  - Kibana: insert `@timestamp` field in the format Kibana expects
  - JSON: attempt to parse the log line as JSON, and if successful set the `ts` field as the log entry time


## Design Considerations

There are a few environmental factors that drove us to this solution:

1. Because we use Kubernetes to run all of our applications, our low level machine configurations are all homogeneous
2. Logging in all programming languages was standardized (JSON, consistent field names)
3. Can't rely on Kubernetes to be running to collect logs (want logs in case something with K8s catches on fire)

### Kubernetes Support

While the log-aggregator does need access to Kubernetes APIs in order to annotate logs from Pods, it is also an invaluable tool in debugging instance startup issues.

Because of this, the log-aggregator (installed as a systemd unit) is started as soon as the instance network is online.
This immediately provides logs that can be reviewed if the instance is having startup issues, without having to SSH to a running instance.

On machines that will run Kubernetes, a configuration file will eventually be written (by the Kubernetes bootstrap process) that tells the kubelet how to talk to the Kubernetes API. The log-aggregator can watch for this file, and as soon as it is detected, it enables the Kubernetes annotation transformer.

### AWS Instance Info

Due to some legacy reasons, instead of reading the instance info from the metadata service, it relies on reading those values from environment variables that are set by another systemd unit that is already pre-installed on our AMIs.


## Why a custom solution vs something like Fluentd?

We actually did run Fluentd in the past. Prior to switching our Docker log driver to Journald, we found that Fluentd was using unusually high CPU to watch the Docker logs directory.

After switching to Journald, we also found that (at the time) the Fluentd/Ruby Journald integration was dropping fields.

We decided to go with a home-built solution once we needed to add JSON parsing and the lazy Kubernetes initializing (described above).

Because our use-case was so straight forward, it was actually simpler to accomplish what we needed in a standalone program, rather than trying to write plugins to fit into the Fluentd architecture.

It also had the bonus of lowering our resource usage.


## Running
There are no cli-flags, all configuration is done via environment variables.

- **FAIR_LOG_CURSOR_PATH**: The path to save the cursor position to
- **FAIR_LOG_FIREHOSE_STREAM**: The Firehose stream name to export to
- **FAIR_LOG_FIREHOSE_CREDENTIALS_ENDPOINT**: (optional) Override the metadata service endpoint to use for credentials
- **FAIR_LOG_K8_CONFIG_PATH**: (optional) The path to watch for the Kubernetes config file
- **FAIR_LOG_K8_CONTAINER_NAME_REGEX**: (optional) Override the built-in regex for extracting the Pod name
- **FAIR_LOG_MOCK_SOURCE**: (optional) Enable a mock source instead of journald (for testing)
- **FAIR_LOG_MOCK_DESTINATION**: (optional) Enable a mock destination (stdout) instead of Kinesis Firehose (for testing)
- **EC2_METADATA_INSTANCE_ID**: (optional) For the AWS transformer
- **EC2_METADATA_LOCAL_IPV4**: (optional) For the AWS transformer
- **EC2_METADATA_LOCAL_HOSTNAME**: (optional) Used by the AWS and the K8s transformer (for the Node name)
- **ENV=production**: (optional) Turns on JSON logging for the aggregator's own logs

## Building/Developing

It requires [dep](https://github.com/golang/dep) for dependencies.

For osx you can run `go build`

For linux, run `make build-linux`, as it requires the systemd development headers.
