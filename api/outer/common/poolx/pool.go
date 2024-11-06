package poolx

import (
	"os"
	"strconv"

	"github.com/panjf2000/ants/v2"
)

var p *ants.Pool

const (
	SYSTEM_POOL_SIZE   = "SYSTEM_POOL_SIZE"
	DOWNLOAD_POOL_SIZE = "DOWNLOAD_POOL_SIZE"
)

func NewSystemPool(options ...ants.Option) (*ants.Pool, error) {
	if p == nil {
		size := 100
		envSize := os.Getenv(SYSTEM_POOL_SIZE)
		if len(envSize) > 0 {
			s, err := strconv.Atoi(envSize)
			if err != nil {
				return nil, err
			}

			size = s
		}

		pl, err := ants.NewPool(size, options...)
		if err != nil {
			return nil, err
		}

		p = pl
		return p, nil
	}

	return p, nil
}

func SystemPoolRelease() {
	if p != nil {
		p.Release()
	}
}

func NewDownloadPool(pf func(interface{}), options ...ants.Option) (*ants.PoolWithFunc, error) {
	size := 10
	envSize := os.Getenv(DOWNLOAD_POOL_SIZE)
	if len(envSize) > 0 {
		s, err := strconv.Atoi(envSize)
		if err != nil {
			return nil, err
		}

		size = s
	}

	pd, err := ants.NewPoolWithFunc(size, pf, options...)
	if err != nil {
		return nil, err
	}

	return pd, nil
}
