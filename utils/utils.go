package utils

import (
	"fmt"
	"math"

	"crypto/sha256"
	"encoding/binary"
)

func Hash(name string, nonce uint64) uint64 {
	hasher := sha256.New()
	hasher.Write([]byte(fmt.Sprintf("%s %d", name, nonce)))
	return binary.BigEndian.Uint64(hasher.Sum(nil))
}

func IsSuccess(err error) {
	if err != nil {
		panic(err)
	}
}

func FindArgminxHash(name string, a uint64, b uint64) uint64 {
	minhashval := uint64(math.Inf(1))
	for x := a; x < b; x++ {
		hashval := Hash(name, x)
		// fmt.Println("Hash of ", x, " is ", hashval)
		if hashval < uint64(minhashval) {
			minhashval = hashval
		}
	}
	return minhashval
}
