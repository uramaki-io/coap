package coap

import "crypto/rand"

const TokenLength = 4

type Token []byte

type TokenSource func() Token

func RandTokenSource(length uint) TokenSource {
	switch {
	case length == 0:
		length = TokenLength
	case length > TokenMaxLength:
		length = TokenMaxLength
	}

	return func() Token {
		token := make(Token, length)
		_, _ = rand.Read(token) // rand.Read never returns an error

		return token
	}
}
