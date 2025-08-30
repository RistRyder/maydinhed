package kafka

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"log"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"
	"github.com/ristryder/maydinhed/mapping"
)

const (
	defaultLocationsTopicName = "locations"
	messageReadTimeout        = 500 * time.Millisecond
	topicCreationTimeout      = 30 * time.Second
)

type Messenger[K mapping.ClusteredMarkerKey] struct {
	consumer          *kafka.Consumer
	consumerCtx       context.Context
	consumerCtxCancel context.CancelFunc
	locationTopicName string
	markerCluster     *mapping.MarkerCluster[K]
	producer          *kafka.Producer
	producerCtx       context.Context
	producerCtxCancel context.CancelFunc
}

func createTopics(kafkaOptions *kafka.ConfigMap, locationTopic *kafka.TopicSpecification) error {
	if locationTopic == nil {
		return nil
	}

	adminClient, adminClientErr := kafka.NewAdminClient(kafkaOptions)
	if adminClientErr != nil {
		return errors.Wrap(adminClientErr, "failed to create Kafka admin client for topic creation")
	}

	ctx, ctxCancel := context.WithTimeout(context.Background(), topicCreationTimeout)

	defer ctxCancel()

	createTopicsResult, createTopicsErr := adminClient.CreateTopics(ctx, []kafka.TopicSpecification{*locationTopic})
	if len(createTopicsResult) == 1 && createTopicsErr == nil {
		return nil
	}

	return errors.Wrap(createTopicsErr, "failed to create location topic")
}

func (m *Messenger[K]) processLocationMessage(message *kafka.Message) error {
	keyedLocation := &mapping.KeyedLocation[K]{}
	if unmarshalErr := json.Unmarshal(message.Value, keyedLocation); unmarshalErr != nil {
		return errors.Wrap(unmarshalErr, "failed to unmarshal location message")
	}

	if markerClusterAddErr := m.markerCluster.Add(*mapping.NewClusteredMarker(*keyedLocation)); markerClusterAddErr != nil {
		return errors.Wrap(markerClusterAddErr, "failed to add marker to cluster")
	}

	return nil
}

func (m *Messenger[K]) startConsumerLoop() {
	go func() {
		for {
			message, messageErr := m.consumer.ReadMessage(messageReadTimeout)
			if messageErr == nil {
				if processLocationErr := m.processLocationMessage(message); processLocationErr != nil {
					log.Printf("failed to process location message from topic '%s': key = %s , value = %s, error = %s\n", *message.TopicPartition.Topic, string(message.Key), string(message.Value), processLocationErr)
				} else {
					log.Printf("successfully processed location message from topic '%s': key = %s , value = %s\n", *message.TopicPartition.Topic, string(message.Key), string(message.Value))
				}
			} else if !messageErr.(kafka.Error).IsTimeout() {
				log.Printf("failed to read Kafka message: %v (%v)\n", messageErr, message)
			}

			select {
			case <-m.consumerCtx.Done():
				log.Println("stopping Kafka consumer")
				return
			default:
				//Carry on
			}
		}
	}()
}

func (m *Messenger[K]) startDeliveryReportHandler() {
	go func() {
		for {
			select {
			case <-m.producerCtx.Done():
				log.Println("stopping Kafka delivery report handler")
				return
			case event := <-m.producer.Events():
				switch typedEvent := event.(type) {
				case *kafka.Message:
					if typedEvent.TopicPartition.Error != nil {
						log.Printf("failed to deliver Kafka message: %v\n", typedEvent.TopicPartition)
					} else {
						log.Printf("successfully delivered Kafka message to %v\n", typedEvent.TopicPartition)
					}
				}
			}
		}
	}()
}

func (m *Messenger[K]) Close() error {
	if m.consumerCtxCancel != nil {
		m.consumerCtxCancel()
	}
	if m.producerCtxCancel != nil {
		m.producerCtxCancel()
	}

	m.producer.Close()
	return m.consumer.Close()
}

func New[K mapping.ClusteredMarkerKey](consumerOptions *kafka.ConfigMap, locationTopic *kafka.TopicSpecification, markerCluster mapping.MarkerCluster[K], producerOptions *kafka.ConfigMap) (*Messenger[K], error) {
	producer, producerErr := kafka.NewProducer(producerOptions)
	if producerErr != nil {
		return nil, errors.Wrap(producerErr, "failed to create Kafka producer")
	}

	if createTopicsErr := createTopics(producerOptions, locationTopic); createTopicsErr != nil {
		return nil, createTopicsErr
	}

	consumerOptions.SetKey("auto.offset.reset", "latest")
	consumerOptions.SetKey("group.id", defaultLocationsTopicName)

	consumer, consumerErr := kafka.NewConsumer(consumerOptions)
	if consumerErr != nil {
		return nil, errors.Wrap(consumerErr, "failed to create Kafka consumer")
	}

	locationTopicName := defaultLocationsTopicName
	if locationTopic != nil {
		locationTopicName = locationTopic.Topic
	}

	if subscribeErr := consumer.SubscribeTopics([]string{locationTopicName}, nil); subscribeErr != nil {
		return nil, errors.Wrap(subscribeErr, "failed to subscribe to locations topic")
	}

	producerCtx, producerCtxCancel := context.WithCancel(context.Background())

	newMessenger := &Messenger[K]{
		consumer:          consumer,
		locationTopicName: locationTopicName,
		markerCluster:     &markerCluster,
		producer:          producer,
		producerCtx:       producerCtx,
		producerCtxCancel: producerCtxCancel,
	}

	defer newMessenger.startDeliveryReportHandler()

	return newMessenger, nil
}

func (m *Messenger[K]) SendLocationUpdate(keyedLocation mapping.KeyedLocation[K]) error {
	locationBytes, locationErr := json.Marshal(keyedLocation)
	if locationErr != nil {
		return errors.Wrap(locationErr, "failed to generate location bytes for update")
	}

	keyBytes := []byte{}
	switch typedKey := any(keyedLocation.Id).(type) {
	case int8:
		keyBytes = []byte{byte(typedKey)}
	case uint8:
		keyBytes = []byte{byte(typedKey)}
	case uint16:
		binary.LittleEndian.PutUint16(keyBytes, typedKey)
	case uint32:
		binary.LittleEndian.PutUint32(keyBytes, typedKey)
	case uint64:
		binary.LittleEndian.PutUint64(keyBytes, typedKey)
	case string:
		keyBytes = []byte(typedKey)
	case uuid.UUID:
		uuidBytes, uuidMarshalErr := typedKey.MarshalBinary()
		if uuidMarshalErr != nil {
			return errors.Wrap(uuidMarshalErr, "failed to generate location key bytes for update")
		}
		keyBytes = uuidBytes
	}

	return m.producer.Produce(&kafka.Message{
		Key: keyBytes,
		TopicPartition: kafka.TopicPartition{
			Partition: kafka.PartitionAny,
			Topic:     &m.locationTopicName,
		},
		Value: locationBytes,
	}, nil)
}

func (m *Messenger[K]) StartListening() error {
	if m.consumerCtxCancel != nil {
		return nil
	}

	if markerClusterStartErr := m.markerCluster.StartAutoClustering(); markerClusterStartErr != nil {
		return errors.Wrap(markerClusterStartErr, "failed to start marker auto clustering")
	}

	m.consumerCtx, m.consumerCtxCancel = context.WithCancel(context.Background())

	m.startConsumerLoop()

	return nil
}

func (m *Messenger[K]) StopListening() error {
	if m.consumerCtxCancel == nil {
		return nil
	}

	if markerClusterStopErr := m.markerCluster.StopAutoClustering(); markerClusterStopErr != nil {
		return errors.Wrap(markerClusterStopErr, "failed to stop marker auto clustering")
	}

	m.consumerCtxCancel()

	m.consumerCtxCancel = nil

	return nil
}
