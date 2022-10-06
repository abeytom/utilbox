package common

import "os"

func IsFile(path string) bool {
	stat, statErr := os.Stat(path)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return false
		}
		panic(statErr)
		return false
	}
	return !stat.IsDir()
}

func IsDir(path string) bool {
	stat, statErr := os.Stat(path)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return false
		}
		panic(statErr)
		return false
	}
	return stat.IsDir()
}

func PathExists(path string) bool {
	_, statErr := os.Stat(path)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return false
		}
		panic(statErr)
		return false
	}
	return true
}
