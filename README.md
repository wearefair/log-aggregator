# Log Aggregator (naming is hard) [![CircleCI](https://circleci.com/gh/wearefair/log-aggregator.svg?style=svg)](https://circleci.com/gh/wearefair/log-aggregator)

Reads logs from Journald, annotates/transforms, and forwards to AWS Kinesis Firehose.

## What is this for?

This is a lightweight process for reading all logs from the system, and forwarding them to a central location for
searching and long-term archival.

It also supports performing minimal transformation and annotation operations on each log entry.

### Is this designed to be generally usable outside of Fair?

Not really, it's engineered to the specifics of our production environments.

Use this code as a reference, or feel free to fork to fit your needs.


### Features/Design

- Input: Journald
- Output: AWS Kinesis Firehose
- Transformations
  - AWS: adds `aws.instance_id`, `aws.local_hostname`, `aws.local_ipv4`
  - Journal: Rename `MESSAGE` field to `log`
  - K8s: Add Pod metadata if the log comes from a Kubernetes Pod
  - Kibana: insert `@timestamp` field in the format Kibana expects
  - Json: attempt to parse the log line as json, and if successful set the `ts` field as the log entry time


## Design Considerations

There are a few environmental factors that drove us to this solution:

1. Because we use Kubernetes to run all of our applications, our low level machine configures are all homogeneous
2. Logging in all programming languages was standardized (JSON, consistent field names)
3. Can't rely on Kubernetes to be running to collect logs

### Kubernetes Support

While the log-aggregator does need access to Kubernetes APIs in order to annotate logs from Pods, it is also an invaluable tool in debugging instance startup issues.

Because of this, the log-aggregator (installed as a Systemd unit) is started as soon as the instance network is online.
This immedietly provides logs that can be reviewed if the instance is having startup issues, without having to SSH to a running instance.

On machines that will run Kubernetes, a configuration file will eventually be written (by the Kubernetes bootstrap process) that tells the kubelet how to talk to the Kubernetes API. The log-aggregator can watch for this file, and as soon as it is detected it enables the Kubernetes annotation transformer.

### AWS Instance Info

Due to some legacy reasons, instead of reading the instance info from the metadata service, it relies on reading those values from environment variables that are set by another Systemd unit that is already preinstalled on our AMIs.


## Why a custom solution vs something like Fluentd?

We actually did run Fluentd in the past. Prior to switching our Docker log driver to Journald, we found that Fluentd was using unusually high CPU watching the docker logs directory.

After switching to Journald, we also found that (at the time) the Fluentd/Ruby Journald integration was dropping fields.

We decided to go with a home-built solution once we needed to add JSON parsing and the lazy Kubernetes initializing (described above).

Because our use-case was so straight forward, it was actually simpler to accomplish what we needed in a standalone program, rather than trying to write plugins to fit into the Fluentd architecture.

It also had the bonus of lowering our resource usage.


## Running
There are no cli-flags, all configuration is done via environment variables.

- **FAIR_LOG_K8_CONFIG_PATH**: (optional) The path to watch for the kubernetes config file
- **FAIR_LOG_K8_CONTAINER_NAME_REGEX**: (optional) override the built-in regex for extracting the pod name
- **FAIR_LOG_CURSOR_PATH**: The path to save the cursor position to.
- **FAIR_LOG_MOCK_SOURCE**: (optional) Enable a mock source instead of journald (for testing)
- **FAIR_LOG_MOCK_DESTINATION**: (optional) Enable a mock destination (stdout) instead of Kinesis Firehose (for testing)
- **FAIR_LOG_FIREHOSE_STREAM**: The firehose stream name to export to
- **FAIR_LOG_FIREHOSE_CREDENTIALS_ENDPOINT**: (optional) override the metadata service endpoint to use for credentials
- **EC2_METADATA_INSTANCE_ID**: (optional) For the aws transformer
- **EC2_METADATA_LOCAL_IPV4**: (optional)For the aws transformer
- **EC2_METADATA_LOCAL_HOSTNAME**: (optional) Used by the AWS and the K8s transformer (for the node name)
- **ENV=production**: (optional) turns on JSON logging for the aggregators own logs

## Building/Developing

It requires dep for dependencies.

For osx you can run `go build`

For linux, run `make build-linux`, as it requires the systemd development headers.
