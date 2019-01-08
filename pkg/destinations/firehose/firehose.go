package firehose

import (
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
	"github.com/wearefair/log-aggregator/pkg/channel"
	"github.com/wearefair/log-aggregator/pkg/logging"
	"github.com/wearefair/log-aggregator/pkg/types"
)

const (
	DefaultBufferFlushLimit = 500
	DefaultFlushInterval    = time.Second * 1

	FirehoseMaxRecords    = 500
	FirehoseMaxRecordSize = 1000 * 1024 // 1000 kb
	FirehoseMaxBatchSize  = 4 * 1024 * 1024
)

type Client struct {
	buffer           <-chan []*types.Record
	progress         chan<- types.Cursor
	firehoseClient   *firehose.Firehose
	firehoseStream   string
	bufferFlushLimit int
	flushInterval    time.Duration
}

type Config struct {
	EC2MetadataEndpoint string
	FirehoseStream      string
	BufferFlushLimit    int
	FlushInterval       time.Duration
}

func New(conf Config) *Client {
	var limit int
	var interval time.Duration

	if conf.BufferFlushLimit == 0 {
		limit = DefaultBufferFlushLimit
	} else {
		limit = conf.BufferFlushLimit
	}

	if conf.FlushInterval == time.Duration(0) {
		interval = DefaultFlushInterval
	} else {
		interval = conf.FlushInterval
	}

	var sess *session.Session
	if conf.EC2MetadataEndpoint != "" {
		sess = awsSession(conf)
	} else {
		sess = session.Must(session.NewSession())
	}

	region := getRegion()
	logging.Logger.Info("Setting aws region", zap.String("region", region))
	client := &Client{
		firehoseClient:   firehose.New(sess, sess.Config.WithRegion(region)),
		firehoseStream:   conf.FirehoseStream,
		bufferFlushLimit: limit,
		flushInterval:    interval,
	}

	return client
}

func (c *Client) Start(records <-chan *types.Record, progress chan<- types.Cursor) {
	if c.buffer != nil {
		panic(errors.New("Tried to start firehose output a second time"))
	}
	c.buffer = channel.NewBufferedChannel(c.bufferFlushLimit, c.flushInterval, records)
	c.progress = progress
	go c.deliver()
}

func (c *Client) deliver() {
	for {
		records, ok := <-c.buffer
		if !ok {
			logging.Logger.Warn("record channel was unexpectedly closed")
			return
		}

		batches := recordsToBatches(records, FirehoseMaxRecords, FirehoseMaxRecordSize, FirehoseMaxBatchSize)

		for _, batch := range batches {
			batchRecords := batch.records

			strategy := backoff.NewExponentialBackOff()
			strategy.MaxElapsedTime = time.Hour * 1
			err := backoff.Retry(func() error {
				input := &firehose.PutRecordBatchInput{
					DeliveryStreamName: aws.String(c.firehoseStream),
					Records:            batchRecords,
				}
				out, err := c.firehoseClient.PutRecordBatch(input)
				if err != nil {
					logging.Logger.Error(fmt.Sprintf("failed tp put record batch: %s", err))
					return err
				}

				if out.FailedPutCount != nil && *out.FailedPutCount != 0 {
					oldBatch := batchRecords
					batchRecords = make([]*firehose.Record, *out.FailedPutCount)
					index := 0
					for i, result := range out.RequestResponses {
						if result.ErrorCode != nil {
							// Skip invalid argument exception, the record is bad for some reason.
							if *result.ErrorCode != firehose.ErrCodeInvalidArgumentException {
								batchRecords[index] = oldBatch[i]
								index++
							}
						}
					}
					// truncate batch in case we don't want to retry all the records.
					batchRecords = batchRecords[:index]
					// return error so that the backoff alg will auto retry the remaining items in the batch.
					err = errors.Errorf("%d items failed to insert, retrying them", *out.FailedPutCount)
					logging.Logger.Error(err.Error())
					return err
				}
				return nil
			}, strategy)

			if err != nil {
				panic(errors.Wrap(err, "Got unrecoverable error publishing to kinesis firehose"))
			}
			// publish batch cursor to progress channel.
			c.progress <- batch.cursor
		}
	}
}

type batch struct {
	cursor  types.Cursor
	records []*firehose.Record
}

func recordsToBatches(records []*types.Record, maxRecords, maxRecordSize, maxBatchSize int) []batch {
	batches := make([]batch, 0)
	current := batch{}
	currentSize := 0

	for _, record := range records {
		serialized, err := json.Marshal(record.Fields)
		if err != nil {
			logging.Error(errors.Wrap(err, "Failed to marshal record to json"))
			continue
		}

		// TODO warn somehow that this is happening.
		if len(serialized) > maxRecordSize-2 {
			serialized = append(serialized[0:maxRecordSize-1], []byte("\n")...)
		} else {
			serialized = append(serialized, []byte("\n")...)
		}

		if currentSize+len(serialized) > maxBatchSize ||
			len(current.records) == maxRecords {
			batches = append(batches, current)
			current = batch{}
			currentSize = 0
		}
		current.records = append(current.records, &firehose.Record{Data: serialized})
		current.cursor = record.Cursor
		currentSize += len(serialized)
	}
	if len(current.records) != 0 {
		return append(batches, current)
	}
	return batches
}

func awsSession(conf Config) *session.Session {
	resolver := func(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
		if service == endpoints.Ec2metadataServiceID {
			return endpoints.ResolvedEndpoint{
				URL:           fmt.Sprintf("http://%s/latest", conf.EC2MetadataEndpoint),
				SigningName:   service,
				SigningMethod: "v4",
			}, nil
		}

		return endpoints.DefaultResolver().EndpointFor(service, region, optFns...)
	}
	return session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config: aws.Config{
			EndpointResolver: endpoints.ResolverFunc(resolver),
		},
	}))
}

func getRegion() string {
	sess := session.Must(session.NewSession())
	if sess.Config.Region != nil && *sess.Config.Region != "" {
		return *sess.Config.Region
	}
	meta := ec2metadata.New(sess)
	if meta.Available() {
		region, err := meta.Region()
		if err == nil {
			return region
		}
	}
	return ""
}
