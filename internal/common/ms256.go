package common

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"unsafe"
)

func NewMultiStep256(seed string) *MultiStep256 {
	m := new(MultiStep256)
	m.encoded = make([]byte, 64)
	m.seed = []byte(seed)
	m.seedLen = len(m.seed)
	m.buff = bytes.NewBuffer(m.seed)
	m.tmp = sha256.Sum256(m.buff.Bytes())
	return m
}

type MultiStep256 struct {
	tmp     [32]byte
	val     string
	encoded []byte
	seed    []byte
	buff    *bytes.Buffer
	seedLen int
}

func (m *MultiStep256) Hash(hashStr string) (hashed string) {
	defer m.Reset()
	m.buff.WriteString(hashStr)
	m.tmp = sha256.Sum256(m.buff.Bytes())
	hex.Encode(m.encoded, m.tmp[:])
	return BytesToString(m.encoded)
}

func (m *MultiStep256) Reset() {
	m.buff.Truncate(m.seedLen)
}

func BytesToString(bytes []byte) string {
	var s string
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bytes))
	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	stringHeader.Data = sliceHeader.Data
	stringHeader.Len = sliceHeader.Len
	return s
}
