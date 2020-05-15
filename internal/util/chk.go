package util

import "slicerapi/internal/logger"

// Chk logs and exits with an error if it's not nil.
func Chk(err error, soft ...bool) {
	if err != nil {
		if len(soft) <= 0 {
			logger.L.Fatalln(err)
		}

		if soft[0] {
			logger.L.Errorln(err)
		}
	}
}
