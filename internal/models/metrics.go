// Package models to hold common models for both services.
package models

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

// Metrics type defines metric that is exchanged between agent and server.
type Metrics struct {
	ID    string   `json:"id"`              // metrics name
	MType string   `json:"type"`            // metrics type: can be counter or gauge
	Delta *int64   `json:"delta,omitempty"` // metrics values if type is counter
	Value *float64 `json:"value,omitempty"` // metrics values if type is gauge
	Hash  string   `json:"hash,omitempty"`  // hash-value
}

// CalcHash calculates hash from key.
func (m *Metrics) CalcHash(key string) {
	m.Hash = m.newHash(key)
}

// CompareHash compares to hashes.
func (m *Metrics) CompareHash(key string) error {
	hash := m.newHash(key)
	if hash != m.Hash {
		err := errors.New("hashes are not equal")
		return err
	}
	return nil
}

// hash generates a hash
func hash(s string, key string) []byte {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(s))
	return h.Sum(nil)
}

// newHash generates new hash for a this metric
func (m *Metrics) newHash(key string) string {
	var h string
	if m.MType == "counter" {
		h = hex.EncodeToString(hash(fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta), key))
	} else {
		h = hex.EncodeToString(hash(fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value), key))
	}
	return h
}
