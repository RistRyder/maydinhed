package kafka

import (
	"context"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/ristryder/maydinhed/stores"
)

const (
	topicCreationTimeout = 30 * time.Second
)

type Messenger[K stores.StoreKey] struct {
	producer *kafka.Producer
}

func createTopics(kafkaOptions *kafka.ConfigMap, topics []kafka.TopicSpecification) error {
	adminClient, adminClientErr := kafka.NewAdminClient(kafkaOptions)
	if adminClientErr != nil {
		return errors.Wrap(adminClientErr, "failed to create Kafka admin client for topic creation")
	}

	ctx, ctxCancel := context.WithTimeout(context.Background(), topicCreationTimeout)

	defer ctxCancel()

	createTopicsResult, createTopicsErr := adminClient.CreateTopics(ctx, topics)
	if createTopicsErr == nil {
		return nil
	}

	createdTopics := []string{}
	for _, createdTopic := range createTopicsResult {
		createdTopics = append(createdTopics, createdTopic.Topic)
	}

	createdTopicsList := strings.Join(createdTopics, ", ")

	return errors.Wrapf(createTopicsErr, "failed to create all topics, only created topics %s (%d of %d)\n", createdTopicsList, len(createdTopics), len(topics))
}

func New[K stores.StoreKey](kafkaOptions *kafka.ConfigMap, topics ...kafka.TopicSpecification) (*Messenger[K], error) {
	producer, producerErr := kafka.NewProducer(kafkaOptions)
	if producerErr != nil {
		return nil, errors.Wrap(producerErr, "failed to create Kafka producer")
	}

	if len(topics) < 1 {
		return &Messenger[K]{
			producer: producer,
		}, nil
	}

	if createTopicsErr := createTopics(kafkaOptions, topics); createTopicsErr != nil {
		return nil, createTopicsErr
	}

	return &Messenger[K]{
		producer: producer,
	}, nil
}

func (m *Messenger[K]) SendLocationUpdate(key K, value stores.Location) error {
	//TODO: Do
	return nil
}
