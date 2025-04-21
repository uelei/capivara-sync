package compressor

import (
	"fmt"
	"github.com/klauspost/compress/zstd"
)

// compressZstd compresses a []byte using zstd
func CompressZstd(data []byte) ([]byte, error) {
	// Create a zstd encoder with default settings
	encoder, err := zstd.NewWriter(nil, zstd.WithEncoderConcurrency(1))
	if err != nil {
		return nil, fmt.Errorf("failed to create zstd encoder: %v", err)
	}
	defer encoder.Close()

	// Compress the data
	compressed := encoder.EncodeAll(data, nil)
	return compressed, nil
}

// decompressZstd decompresses a zstd-compressed []byte
func DecompressZstd(compressed []byte) ([]byte, error) {
	decoder, err := zstd.NewReader(nil, zstd.WithDecoderConcurrency(1))
	if err != nil {
		return nil, fmt.Errorf("failed to create zstd decoder: %v", err)
	}
	defer decoder.Close()
	return decoder.DecodeAll(compressed, nil)
}
