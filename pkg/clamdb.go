package btcplex

import (
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
)

type DigInfo struct {
	Total  uint64 `json:"total"`
	Dug    uint64 `json:"dug"`
	Buried uint64 `json:"buried"`
}

type ClamSpeech struct {
	Comment string `json:"comment"`
	Height  uint   `json:"height"`
	Index   uint32 `json:"tx_index"`
}

// GetSpeechesAt returns the speeches for txs at a given height.
func GetSpeechesAt(rpool *redis.Pool, height uint) (speeches map[uint32]string, err error) {
	c := rpool.Get()
	defer c.Close()
	v, err := redis.Values(c.Do("HGETALL", fmt.Sprintf("speech:%v", height)))
	if err != nil {
		return nil, err
	}

	speeches = make(map[uint32]string)
	for i := 0; i < len(v); i += 2 {
		tmp_tx_index, _ := redis.Uint64(v[i], nil)
		tx_index := uint32(tmp_tx_index)

		comment, _ := redis.String(v[i+1], nil)
		speeches[tx_index] = comment
	}

	return
}

// GetSpeeches returns recent ClamSpeeches.
func GetSpeeches(rpool *redis.Pool, blockCount int) (speeches []*ClamSpeech, startHeight uint, err error) {
	c := rpool.Get()
	defer c.Close()

	lastheight, err := redis.Int(c.Do("GET", "height:latest"))
	if err != nil {
		return nil, 0, err
	}
	if lastheight-blockCount <= 0 {
		return nil, 0, errors.New("Count is higher than the number of blocks.")
	}
	startHeight = uint(lastheight - blockCount)

	speeches = []*ClamSpeech{}
	for i := 0; i < blockCount; i++ {
		height := lastheight - i

		blockSpeeches, err := GetSpeechesAt(rpool, uint(height))
		if err != nil {
			continue
		}
		for tx_index, comment := range blockSpeeches {
			speech := &ClamSpeech{
				Height:  uint(height),
				Comment: comment,
				Index:   tx_index,
			}
			speeches = append(speeches, speech)
		}

	}

	return
}

// GetDigInfo returns data on dug clams.
func GetDigInfo(rpool *redis.Pool) (info *DigInfo, err error) {
	c := rpool.Get()
	defer c.Close()

	info = new(DigInfo)
	info.Total = 3208032 * 460545574
	undugCnt, _ := redis.Int64(c.Do("SCARD", "undug"))
	info.Buried = uint64(undugCnt) * 460545574
	info.Dug = info.Total - info.Buried

	return
}
