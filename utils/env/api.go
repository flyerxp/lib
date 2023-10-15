package env

import "os"

func GetEnv() string {
	e := os.Getenv("GO_ENV")
	if len(e) == 0 {
		return "test"
	}
	return e
}
func GetConfRoot() string {
	e := os.Getenv("CONF_ROOT")
	if len(e) == 0 {
		return ""
	}
	return e
}
