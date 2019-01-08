// Reads ec2 metadata instance environment variables and sets them on the logs.
package aws

import (
	"os"

	"github.com/wearefair/log-aggregator/pkg/transform"
	"github.com/wearefair/log-aggregator/pkg/types"
)

// These are set by instance-environment.service, which is a standard process
// on all our base AMIs.

const (
	EnvInstanceId    = "EC2_METADATA_INSTANCE_ID"
	EnvLocalIpv4     = "EC2_METADATA_LOCAL_IPV4"
	EnvLocalHostname = "EC2_METADATA_LOCAL_HOSTNAME"
)

func New() transform.Transformer {
	meta := metadata{
		InstanceId:    os.Getenv(EnvInstanceId),
		LocalIpv4:     os.Getenv(EnvLocalIpv4),
		LocalHostname: os.Getenv(EnvLocalHostname),
	}

	return func(rec *types.Record) (*types.Record, error) {
		rec.Fields["aws"] = meta
		return rec, nil
	}
}

type metadata struct {
	InstanceId     string `json:"instance_id,omitempty"`
	InstanceType   string `json:"instance_type,omitempty"`
	LocalHostname  string `json:"local_hostname,omitempty"`
	LocalIpv4      string `json:"local_ipv4,omitempty"`
	PublicHostname string `json:"public_hostname,omitempty"`
	PublicIpv4     string `json:"public_ipv4,omitempty"`
}
