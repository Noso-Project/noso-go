package common

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"reflect"
	"strings"
	"unsafe"

	_ "github.com/minio/sha256-simd"
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
	m.seedLen = len(seed)
	m.buff = bytes.NewBuffer(m.seed)
	m.buffLen = m.buff.Len()
	// m.tmp = sha256.Sum256(m.buff.Bytes())
	return m
}

type MultiStep256 struct {
	HashedString string
	tmp          [32]byte
	val          string
	encoded      []byte
	seed         []byte
	buff         *bytes.Buffer
	buffLen      int
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

func (m *MultiStep256) HashTest(h []byte) (hashed string) {
	m.Reset()
	m.buff.Write(h)
	m.tmp = sha256.Sum256(m.buff.Bytes())
	hex.Encode(m.encoded, m.tmp[:])
	m.HashedString = BytesToString(m.encoded)
	return m.HashedString
}

func (m *MultiStep256) Search(targets []string) (match string) {
	var target string
	// fmt.Println("Targets are:", targets)
	if !strings.Contains(m.HashedString, targets[0]) {
		return
	}

	match = targets[0]
	for _, target = range targets[1:] {
		if !strings.Contains(m.HashedString, target) {
			return
		}
		match = target
	}
	return
}

// TODO: This might not be needed
func (m *MultiStep256) Reset() {
	m.buff.Truncate(m.buffLen)
}

func MultiStep256SumHash(val string) []byte {
	return NewMultiStep256Sum(val).Hash(nil)
}

func NewMultiStep256Sum(seed string) *MultiStep256Sum {
	m := new(MultiStep256Sum)
	m.out = make([]byte, 64)
	m.seed = []byte(seed)
	m.hash = sha256.New()
	return m
}

type MultiStep256Sum struct {
	out  []byte
	seed []byte
	hash hash.Hash
}

func (m *MultiStep256Sum) Hash(hashBytes []byte) []byte {
	m.hash.Reset()
	m.hash.Write(m.seed)
	m.hash.Write(hashBytes)
	m.hash.Sum(m.out)
	return m.out
}

func (m *MultiStep256Sum) Search(targets [][]byte) (match []byte) {
	var target []byte
	if !bytes.Contains(m.out, targets[0]) {
		return
	}

	match = targets[0]
	for _, target = range targets[1:] {
		if !bytes.Contains(m.out, target) {
			return
		}
		match = target
	}
	return
}

func BytesToString(bytes []byte) string {
	var s string
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bytes))
	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	stringHeader.Data = sliceHeader.Data
	stringHeader.Len = sliceHeader.Len
	return s
}
