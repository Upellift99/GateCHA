package altcha

import (
	"time"

	lib "github.com/altcha-org/altcha-lib-go"
)

func GenerateChallenge(hmacSecret string, maxNumber int64, algorithm string, expireSeconds int) (lib.Challenge, error) {
	expires := time.Now().Add(time.Duration(expireSeconds) * time.Second)
	opts := lib.ChallengeOptions{
		HMACKey:   hmacSecret,
		MaxNumber: maxNumber,
		Algorithm: lib.Algorithm(algorithm),
		Expires:   &expires,
	}
	return lib.CreateChallenge(opts)
}

func VerifyPayload(hmacSecret string, payload string) (bool, error) {
	return lib.VerifySolution(payload, hmacSecret, true)
}
