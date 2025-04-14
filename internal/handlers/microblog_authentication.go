package handlers

import (
	"os"
)

const (
	MicroblogToken = "MICROBLOG_AUTH_TOKEN"
)

func IsAuthenticated(password string) bool {

	return password == os.Getenv(MicroblogToken)
}
