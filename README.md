# Log Aggregator

Reads logs from journald, annotates, and forwards them to our configured location(s).

## Building

For osx you can run `go build`

For linux, run `make build-linux`, as it requires the systemd development headers.
