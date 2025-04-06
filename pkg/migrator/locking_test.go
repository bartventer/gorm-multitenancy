package migrator

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateLockKey(t *testing.T) {
	tests := []struct {
		name string
		text string
		want uint64
	}{
		{
			name: "test",
			text: "test",
			want: 18007334074686647077,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, GenerateLockKey(tt.text))
		})
	}
}

func GenerateLockKeyMD5(s string) uint64 {
	hash := md5.Sum([]byte(s))
	bigInt := new(big.Int).SetBytes(hash[:])
	return bigInt.Uint64()
}

func GenerateLockKeySha1(s string) uint64 {
	hash := sha1.Sum([]byte(s))
	bigInt := new(big.Int).SetBytes(hash[:])
	return bigInt.Uint64()
}
func GenerateLockKeySha256(s string) uint64 {
	hash := sha256.Sum256([]byte(s))
	bigInt := new(big.Int).SetBytes(hash[:])
	return bigInt.Uint64()
}

func GenerateLockKeySha512(s string) uint64 {
	hash := sha512.Sum512([]byte(s))
	bigInt := new(big.Int).SetBytes(hash[:])
	return bigInt.Uint64()
}

var result uint64

func BenchmarkGenerateLockKey(b *testing.B) {
	for _, f := range []struct {
		name string
		fn   func(string) uint64
	}{
		{name: "FNV-1a", fn: GenerateLockKey},
		{name: "MD5", fn: GenerateLockKeyMD5},
		{name: "Sha1", fn: GenerateLockKeySha1},
		{name: "Sha256", fn: GenerateLockKeySha256},
		{name: "Sha512", fn: GenerateLockKeySha512},
	} {
		b.Run(f.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				result = f.fn("test")
			}
		})
	}
}
