package random

import (
	"encoding/binary"
	uuid "github.com/satori/go.uuid"
)

// UUIDGenerator implements UIntRandomizer
// that creates random unsigned integer using UUID package
type UUIDGenerator struct {
	version string
}

// NewUUIDGenerator creates new UUIDGenerator
func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{version: "v4"}
}

// Uint64 generates random uint64 number
func (gen *UUIDGenerator) Uint64() uint64 {
	var u = uuid.NewV4()
	return binary.BigEndian.Uint64(u[:8])
}

// Uint32 generates random uint32 number
func (gen *UUIDGenerator) Uint32() uint32 {
	var u = uuid.NewV4()
	return binary.BigEndian.Uint32(u[:4])
}

