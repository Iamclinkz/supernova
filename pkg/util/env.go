package util

import "os"

func GetEnv() string {
	if os.Getenv("env") == "k8s" {
		return "k8s"
	}

	return "dev"
}
