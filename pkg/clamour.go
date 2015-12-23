package btcplex

import (
	"encoding/hex"
	"errors"
	"github.com/garyburd/redigo/redis"
	"strings"
)

// ClamourInfo models the status of recent petitions.
type ClamourInfo struct {
	StartBlock uint              `json:"start_block"`
	EndBlock   uint              `json:"end_block"`
	Petitions  map[string][]uint `json:"petitions"`
}

// GetClamourInfo returns the status of CLAMour petitions.
func GetClamourInfo(rpool *redis.Pool) (info *ClamourInfo, err error) {
	speeches, startHeight, err := GetSpeeches(rpool, 10000)
	if err != nil {
		return nil, err
	}

	c := rpool.Get()
	defer c.Close()

	info = &ClamourInfo{
		StartBlock: startHeight,
		EndBlock:   startHeight + 10000,
		Petitions:  make(map[string][]uint),
	}

	for _, v := range speeches {
		// Only coinstake speeches count.
		if v.Index != 1 {
			continue
		}
		pids, err := ParseClamourSpeech(v.Comment)
		if err != nil {
			continue
		}
		for _, pid := range pids {
			info.Petitions[pid] = append(info.Petitions[pid], v.Height)
		}
	}

	return
}

// ParseClamourSpeech parses a clamour clamspeech into a list of petition IDs.
func ParseClamourSpeech(speech string) (pids []string, err error) {
	if !strings.HasPrefix(speech, "clamour ") {
		return nil, errors.New("Non-clamour speech")
	}
	speech = strings.TrimPrefix(speech, "clamour ")
	strs := strings.Split(speech, " ")
	pids = []string{}

	for _, v := range strs {
		if _, err = hex.DecodeString(v); err != nil {
			continue
		}
		if len(v) != 8 {
			continue
		}
		pids = append(pids, v)
	}

	return
}
