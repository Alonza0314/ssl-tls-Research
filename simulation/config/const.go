package config

import (
	"time"

	"golang.org/x/crypto/chacha20poly1305"
)

const (
	CLIENT_HELLO = "CLIENT_HELLO"
	SERVER_HELLO = "SERVER_HELLO"
	TX_HEADER    = "txHeader"

	BUFFERSIZE = 1024

	X25519_KEY_SIZE = 32

	SEED_SIZE        = 32
	SECERT_KEY_SIZE  = 32
	PUBLIC_KEY_SIZE  = 32
	SESSION_KEY_SIZE = 32

	STREAM_KEY_SIZE    = chacha20poly1305.KeySize
	STREAM_HEADER_SIZE = chacha20poly1305.NonceSizeX

	CRYPTO_CORE_HCHACHA20_INPUTSIZE                    = 16
	CRYPTO_SECRETSTREAM_XCHACHA20POLY1305_COUNTERBYTES = 4

	CONN_TIMEOUT = 5 * time.Second

	NEW      = "[NEW]"
	RECEIVED = "[RECEIVED]"
	SENT     = "[SENT]"
)
