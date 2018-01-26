package token

import (
	"bufio"
	"crypto/rand"
	"fmt"
)

// This is lifted directly from kubeadm
// I didn't want to have to import all k8s just for one funcition
// source: https://github.com/kubernetes/kubernetes/blob/master/cmd/kubeadm/app/util/token/tokens.go

const (
	// TokenIDBytes defines a number of bytes used for a token id
	TokenIDBytes = 6
	// TokenSecretBytes defines a number of bytes used for a secret
	TokenSecretBytes = 16
	// valid chars for a bootstrap token
	validBootstrapTokenChars = "0123456789abcdefghijklmnopqrstuvwxyz"
)

func randBytes(length int) (string, error) {
	// len("0123456789abcdefghijklmnopqrstuvwxyz") = 36 which doesn't evenly divide
	// the possible values of a byte: 256 mod 36 = 4. Discard any random bytes we
	// read that are >= 252 so the bytes we evenly divide the character set.
	const maxByteValue = 252

	var (
		b     byte
		err   error
		token = make([]byte, length)
	)

	reader := bufio.NewReaderSize(rand.Reader, length*2)
	for i := range token {
		for {
			if b, err = reader.ReadByte(); err != nil {
				return "", err
			}
			if b < maxByteValue {
				break
			}
		}

		token[i] = validBootstrapTokenChars[int(b)%len(validBootstrapTokenChars)]
	}

	return string(token), nil
}

// GenerateToken generates a new token with a token ID that is valid as a
// Kubernetes DNS label.
// For more info, see kubernetes/pkg/util/validation/validation.go.
func GenerateToken() (string, error) {
	tokenID, err := randBytes(TokenIDBytes)
	if err != nil {
		return "", err
	}

	tokenSecret, err := randBytes(TokenSecretBytes)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s.%s", tokenID, tokenSecret), nil
}
