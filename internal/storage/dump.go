package storage

import (
	"encoding/json"
	"os"

	"github.com/go-errors/errors"
)

// producer type to write json to file.
type producer struct {
	file    *os.File
	encoder *json.Encoder
}

// NewProducer creates new producer.
func NewProducer(filename string) (*producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, errors.Wrap(err, 1)
	}
	return &producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

// Close closes file.
func (p *producer) Close() error {
	return p.file.Close()
}

// consumer type to read from json file.
type consumer struct {
	file    *os.File
	decoder *json.Decoder
}

// NewConsumer creates new consumer.
func NewConsumer(fileName string) (*consumer, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, errors.Wrap(err, 1)
	}
	return &consumer{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *consumer) Close() error {
	return c.file.Close()
}
