package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cast"
	"time"

	"github.com/elwin/franz/pkg/franz"

	"github.com/spf13/cobra"
)

func init() {
	var (
		partitions []int
		count      int64
		start      string
		duration   time.Duration
		follow     bool
		decode     bool
	)

	var monitorCmd = &cobra.Command{
		Use:     "consume [topic]",
		Aliases: []string{"monitor"},
		Short:   "Consume a specific kafka topic",
		Long: `Consume a specific kafka topic.

You may consume from a kafka topic with arbitrary offsets.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			topic := args[0]

			return execute(func(ctx context.Context, f *franz.Franz) (string, error) {
				if follow && duration > 0 {
					return "", errors.New("when \"follow\" is enabled, duration may not be set")
				}

				if count > 0 && duration > 0 {
					return "", errors.New("only one of \"count\" or \"duration\" may be used")
				}

				var from time.Time
				if start != "" {
					var err error
					from, err = cast.StringToDate(start)
					if err != nil {
						return "", err
					}
				}

				var to time.Time
				if duration > 0 {
					if start == "" {
						from = time.Now().Add(-duration)
					} else {
						to = from.Add(duration)
					}
				}

				req := franz.ConsumeRequest{
					Topic:      topic,
					From:       from,
					To:         to,
					Count:      count,
					Follow:     follow,
					Partitions: convertSliceIntToInt32(partitions),
					Decode:     decode,
				}

				messages, err := f.HistoryEntries(req, ctx)
				if err != nil {
					return "", err
				}

				for message := range messages {
					out, err := formatJSON(message)
					if err != nil {
						return "", err
					}

					fmt.Println(out)
				}

				return "", nil
			})
		},
	}

	RootCmd.AddCommand(monitorCmd)

	monitorCmd.Flags().Int64VarP(&count, "count", "n", 0, "Consumes the n last messages for each partition")
	monitorCmd.Flags().IntSliceVarP(&partitions, "partitions", "p", nil, "The partitions to consume (comma-separated), all partitions will be used if not set")
	monitorCmd.Flags().DurationVarP(&duration, "duration", "d", 0, "Time-frame after \"start\", or if no start is defined, use the specified duration before now as start; may not be used with \"follow\"")
	monitorCmd.Flags().StringVarP(&start, "start", "s", "", "Starting time")
	monitorCmd.Flags().BoolVarP(&follow, "follow", "f", false, "Consume future messages when they arrive")
	monitorCmd.Flags().BoolVar(&decode, "decode", false, "Decodes the message according to the schema defined in the schema registry")
}
