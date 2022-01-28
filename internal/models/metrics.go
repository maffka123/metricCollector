package models

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

type Metrics struct {
	ID    string   `json:"id"`              // metrics name
	MType string   `json:"type"`            // metrics type: can be counter or gauge
	Delta *int64   `json:"delta,omitempty"` // metrics values if type is counter
	Value *float64 `json:"value,omitempty"` // metrics values if type is gauge
	Hash  string   `json:"hash,omitempty"`  // hash-value
}

func (m *Metrics) CalcHash(key string) {
	m.Hash = m.newHash(key)
}

func (m *Metrics) CompareHash(key string) error {
	hash := m.newHash(key)
	if hash != m.Hash {
		err := errors.New("hashes are not equal")
		return err
	}
	return nil
}

func hash(s string, key string) []byte {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(s))
	return h.Sum(nil)
}

func (m *Metrics) newHash(key string) string {
	var h string
	if m.MType == "counter" {
		h = hex.EncodeToString(hash(fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta), key))
	} else {
		h = hex.EncodeToString(hash(fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value), key))
	}
	return h
}
