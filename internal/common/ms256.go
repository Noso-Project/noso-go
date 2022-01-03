package common

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"reflect"
	"unsafe"
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

func NewFoo(seed string) *Foo {
	f := new(Foo)
	f.hasher = sha256.New()
	f.hasher.Write([]byte(seed))
	return f
}

type Foo struct {
	hasher hash.Hash
}

func (f *Foo) Hash(hashStr string) []byte {
	return f.hasher.Sum([]byte(hashStr))
}

func NewBar(seed string) *Bar {
	b := new(Bar)
	b.hasher = sha256.New()
	b.hasher.Write([]byte(seed))
	return b
}

type Bar struct {
	hasher hash.Hash
}

func (b *Bar) Hash(hashStr string) []byte {
	return b.hasher.Sum([]byte(hashStr))
}

func (b *Bar) HashBytes(hashBytes []byte) []byte {
	return b.hasher.Sum(hashBytes)
}

func BytesToString(bytes []byte) string {
	var s string
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&bytes))
	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	stringHeader.Data = sliceHeader.Data
	stringHeader.Len = sliceHeader.Len
	return s
}
