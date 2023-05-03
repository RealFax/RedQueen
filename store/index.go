package store

import (
	"crypto/sha256"
	"github.com/RealFax/RedQueen/utils"
)

const KeyIndexNamespace string = "_KeyIndex"

type KeyIndex struct {
	NodeID    int64  `msgpack:"nid"`
	Namespace string `msgpack:"ns"`
	Key       []byte `msgpack:"key"`
}

func (i *KeyIndex) FullPath() []byte {
	return append(utils.String2Bytes(i.Namespace), i.Key...)
}

func (i *KeyIndex) Hash() []byte {
	s := sha256.New()
	s.Write(i.FullPath())
	return s.Sum(nil)
}

func NewKeyIndex(nodeID int64, namespace string, key []byte) *KeyIndex {
	return &KeyIndex{
		NodeID:    nodeID,
		Namespace: namespace,
		Key:       key,
	}
}
