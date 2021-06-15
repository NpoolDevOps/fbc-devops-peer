package baseapi

import (
	"os"

	log "github.com/EntropyPool/entropy-logger"
)

func GetFileIfWriteRead(file string) (float64, error) {
	fi, err := os.Lstat(file)
	if err != nil {
		log.Errorf(log.Fields{}, "err is:", err)
		return 0, err
	}
	if (fi.Mode().Perm() | 0755) > 0 {
		return 1, nil
	} else {
		return 0, nil
	}
}
