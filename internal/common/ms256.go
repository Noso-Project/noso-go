package common

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"strings"
	"unsafe"
)

const (
	HashableSeedChars = "!\"#$&')*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^`abcdefghijklmnopqrstuvwxyz{|"
)

func MultiStep256Hash(val string) string {
	return NewMultiStep256(val).Hash("")
}

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
	HashedString string
	tmp          [32]byte
	val          string
	encoded      []byte
	seed         []byte
	buff         *bytes.Buffer
	seedLen      int
}

func (m *MultiStep256) Hash(hashStr string) (hashed string) {
	defer m.Reset()
	m.buff.WriteString(hashStr)
	m.tmp = sha256.Sum256(m.buff.Bytes())
	hex.Encode(m.encoded, m.tmp[:])
	m.HashedString = BytesToString(m.encoded)
	return m.HashedString
}

func (m *MultiStep256) Search(targets []string) (match string) {
	// fmt.Println("Targets are:", targets)
	if !strings.Contains(m.HashedString, targets[0]) {
		return
	}
	for _, target := range targets {
		if strings.Contains(m.HashedString, target) {
			match = target
		}
	}
	return
}

// TODO: This might not be needed
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
