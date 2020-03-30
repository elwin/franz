package franz

import (
	"testing"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/require"
)

func TestDiffTopics(t *testing.T) {
	topicA := Topic{Name: "testA", NumPartitions: 1, ReplicationFactor: 1}
	topicB := Topic{Name: "testB", NumPartitions: 1, ReplicationFactor: 1}
	topicC := Topic{Name: "testC", NumPartitions: 1, ReplicationFactor: 1}

	existingTopics := []Topic{topicA, topicC}
	newTopics := []Topic{topicB, topicA}

	toCreate, toDelete, toAlter := diffTopics(newTopics, existingTopics)

	require.Len(t, toCreate, 1)
	require.Contains(t, toCreate, topicB)

	require.Len(t, toDelete, 1)
	require.Contains(t, toDelete, topicC)

	require.Len(t, toAlter, 0)
}

func TestDiffTopicsConfig(t *testing.T) {
	existingTopics := []Topic{{Name: "test", Configs: []sarama.ConfigEntry{{Name: "log_compaction", Value: "true"}, {Name: "log.retention.ms", Value: "100"}}}}
	newTopics := []Topic{{Name: "test", Configs: []sarama.ConfigEntry{{Name: "log_compaction", Value: "false"}, {Name: "log.retention.ms", Value: "200"}}}}

	toCreate, toDelete, toAlter := diffTopics(newTopics, existingTopics)

	require.Len(t, toCreate, 0)
	require.Len(t, toDelete, 0)
	require.Len(t, toAlter, 1)
	require.Equal(t, toAlter[0].TopicName, "test")
	require.Equal(t, *toAlter[0].Configs["log_compaction"], "false")
	require.Equal(t, *toAlter[0].Configs["log.retention.ms"], "200")
}

func TestTopicValidity(t *testing.T) {
	require.NoError(t, validateTopics([]Topic{{Name: "valid"}}))
	require.Error(t, validateTopics([]Topic{{Name: "_invalid"}}))
}

func TestDiffTopicsConfigDefaultConfig(t *testing.T) {
	existingTopics := []Topic{{
		Name:              "test",
		NumPartitions:     1,
		ReplicationFactor: 1,
	}}

	newTopics := []Topic{{
		Name:              "test",
		NumPartitions:     1,
		ReplicationFactor: 1,
		Configs:           []sarama.ConfigEntry{{Name: "retention.ms", Value: "2"}},
	}}

	toCreate, toDelete, toAlter := diffTopics(newTopics, existingTopics)

	require.Len(t, toCreate, 0)
	require.Len(t, toDelete, 0)
	require.Len(t, toAlter, 1)
	require.Equal(t, *(toAlter[0].Configs["retention.ms"]), "2")
}
