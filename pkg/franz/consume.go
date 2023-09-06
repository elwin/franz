package franz

import (
	"context"
	"github.com/IBM/sarama"
	"golang.org/x/sync/errgroup"
	"io"
	"time"
)

type Result struct {
	err error
	msg Message
}

type Receiver struct {
	cancel             context.CancelFunc
	ctx                context.Context
	messageC           chan Result
	availableConsumers int
}

type MonitorRequest struct {
	Topic      string
	Partitions []int32
	Count      int64
	Follow     bool
	Decode     bool
}

type ConsumeRequest struct {
	Topic      string
	From, To   time.Time // To takes precedence over Count
	Count      int64
	Follow     bool
	Partitions []int32
	Decode     bool
}

// Stop instructs the receiver to finish receiving messages.
// Does not block until all goroutines have finished. Instead,
// Next() can be called to drain the remaining messages and
// wait until all goroutines have finished.
func (r *Receiver) Stop() {
	r.cancel()
}

// Next retrieves the next message. If there are no more messages,
// io.EOF is returned. Not thread-safe.
func (r *Receiver) Next() (Message, error) {
	for result := range r.messageC {
		if result.err == io.EOF {
			r.availableConsumers--
			if r.availableConsumers == 0 {
				return Message{}, io.EOF
			}
			continue
		} else if result.err != nil {
			return Message{}, result.err
		}

		return result.msg, nil
	}

	return Message{}, io.EOF
}

func (f *Franz) HistoryEntries(req ConsumeRequest, ctx context.Context) (chan Message, error) {
	if len(req.Partitions) == 0 {
		p, err := f.client.Partitions(req.Topic)
		if err != nil {
			return nil, err
		}

		req.Partitions = p
	}

	consumer, err := sarama.NewConsumerFromClient(f.client)
	if err != nil {
		return nil, err
	}
	defer consumer.Close()

	messages := make(chan Message)

	errs, _ := errgroup.WithContext(ctx)

	for _, partition := range req.Partitions {
		partition := partition

		errs.Go(func() error {
			startOffset, err := f.client.GetOffset(req.Topic, partition, req.From.UnixNano()/int64(time.Millisecond))
			if err != nil {
				return err
			}

			if startOffset == sarama.OffsetNewest {
				f.log.Warnf("no messages (yet) available on partition %d, starting with offset 0", partition)
				startOffset = 0
			}

			// endOffset refers to the message that will be produced next,
			// i.e. we stop at the message received one before.
			var endOffset int64
			if !req.To.Equal(time.Time{}) {
				// Use absolute time

				offset, err := f.client.GetOffset(req.Topic, partition, req.To.UnixNano()/int64(time.Millisecond))
				if err != nil {
					return err
				}

				endOffset = offset
			} else if req.Count > 0 {
				// Use offset count

				endOffset = startOffset + req.Count

				latestOffset, err := f.client.GetOffset(req.Topic, partition, sarama.OffsetNewest)
				if err != nil {
					return err
				}

				if !req.Follow && endOffset > latestOffset {
					endOffset = latestOffset
				}
			} else if req.Follow {
				// Just continue until the user cancels

				endOffset = -1 // never end
			} else {
				// Use the latest offset

				offset, err := f.client.GetOffset(req.Topic, partition, sarama.OffsetNewest)
				if err != nil {
					return err
				}

				endOffset = offset
			}

			partitionConsumer, err := consumer.ConsumePartition(req.Topic, partition, startOffset)
			if err != nil {
				return err
			}

			messageC := partitionConsumer.Messages()

		loop:
			for {
				select {
				case <-ctx.Done():
					break loop

				case message := <-messageC:
					messageTransformed := Message{
						Topic:     message.Topic,
						Timestamp: message.Timestamp,
						Partition: message.Partition,
						Key:       string(message.Key),
						Value:     string(message.Value),
						Offset:    message.Offset,
					}

					if req.Decode {
						decoded, err := f.codec.Decode(message.Value)
						if err != nil {
							return err
						}

						messageTransformed.Value = string(decoded)
					}

					messages <- messageTransformed

					if (message.Offset + 1) == endOffset {
						break loop
					}
				}
			}

			if err := partitionConsumer.Close(); err != nil {
				f.log.Error(err)
			}

			return nil
		})
	}

	go func() {
		if err := errs.Wait(); err != nil {
			f.log.Error(err)
		}

		close(messages)
	}()

	return messages, nil
}
