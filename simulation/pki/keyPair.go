package pki

import (
	"crypto/rand"
	"errors"
	"fmt"
	"simulation/config"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/curve25519"
)

// type KeyPair struct {
// 	privateKey ed25519.PrivateKey
// 	publicKey  ed25519.PublicKey
// }

// func NewKeyPair(private, public string) (*KeyPair, error) {
// 	privatePem, err := os.ReadFile(private)
// 	if err != nil {
// 		return nil, errors.New("failed to read privatePem: " + err.Error())
// 	}
// 	privateBlock, _ := pem.Decode(privatePem)
// 	if privateBlock == nil || privateBlock.Type != "PRIVATE KEY" {
// 		return nil, errors.New("failed to decode privatePem")
// 	}
// 	privateStruct, err := x509.ParsePKCS8PrivateKey(privateBlock.Bytes)
// 	if err != nil {
// 		return nil, errors.New("failed to parse privateBlock: " + err.Error())
// 	}
// 	privateKey, ok := privateStruct.(*ecdh.PrivateKey)
// 	if !ok {
// 		return nil, errors.New("failed to assert privateStruct")
// 	}

// 	publicPem, err := os.ReadFile(public)
// 	if err != nil {
// 		return nil, errors.New("failed to read publicPem:" + err.Error())
// 	}
// 	publicBlock, _ := pem.Decode(publicPem)
// 	if publicBlock == nil || publicBlock.Type != "PUBLIC KEY" {
// 		return nil, errors.New("failed to decode publicPem")
// 	}
// 	publicStruct, err := x509.ParsePKIXPublicKey(publicBlock.Bytes)
// 	if err != nil {
// 		return nil, errors.New("failed to parse publicBlock: " + err.Error())
// 	}
// 	publicKey, ok := publicStruct.(*ecdh.PublicKey)
// 	if !ok {
// 		return nil, errors.New("failed to assert publicStruct")
// 	}

// 	k := KeyPair{}
// 	k.privateKey, k.publicKey = privateKey.Bytes(), publicKey.Bytes()

// 	return &k, nil
// }

// func (k *KeyPair) Private() ed25519.PrivateKey {
// 	return k.privateKey
// }

// func (k *KeyPair) Public() ed25519.PublicKey {
// 	return k.publicKey
// }

// func (k *KeyPair) SessionKeys(peerKey ed25519.PublicKey) ([]byte, []byte, error) {
// 	q, err := curve25519.X25519(k.Private(), peerKey)
// 	if err != nil {
// 		return nil, nil, errors.New("failed to compute share point in x25519:\n\t" + err.Error())
// 	}

// 	h, err := blake2b.New(2 * config.SESSION_KEY_SIZE, nil)
// 	if err != nil {
// 		return nil, nil, errors.New("failed to new blake2b:\n\t" + err.Error())
// 	}

// 	for _, b := range [][]byte{q, k.Public(), peerKey} {
// 		if _, err := h.Write(b); err != nil {
// 			return nil, nil, errors.New("failed to write to hash:\n\t" + err.Error())
// 		}
// 	}

// 	keys := h.Sum(nil)

// 	return keys[: config.SESSION_KEY_SIZE], keys[config.SESSION_KEY_SIZE:], nil
// }

type KeyPair struct {
	pk [config.PUBLIC_KEY_SIZE]byte
	sk [config.SESSION_KEY_SIZE]byte
}

func NewKeyPair() (*KeyPair, error) {
	seed := make([]byte, config.SEED_SIZE)
	if _, err := rand.Read(seed); err != nil {
		return nil, errors.New("failed to new key pair:\n\t" + err.Error())
	}
	return newKeyPairFromSeed(seed)
}

func newKeyPairFromSeed(seed []byte) (*KeyPair, error) {
	kp := &KeyPair{}

	hash, _ := blake2b.New(config.SECERT_KEY_SIZE, nil)
	hash.Write(seed)
	sk := hash.Sum(nil)
	if len(sk) != config.SECERT_KEY_SIZE {
		return nil, errors.New("failed to hash from seed:\n\t" + errors.ErrUnsupported.Error())
	}
	copy(kp.sk[:], sk)

	pk, err := curve25519.X25519(kp.sk[:], curve25519.Basepoint)
	if err != nil {
		return nil, errors.New("failed to get pk from basepoint:\n\t" + err.Error())
	}
	if len(pk) != config.PUBLIC_KEY_SIZE {
		return nil, errors.New("failed to get pk from basepoint:\n\tpk size expected to be " + fmt.Sprintf("%v", config.PUBLIC_KEY_SIZE) + " but " + fmt.Sprintf("%v", len(pk)))
	}
	copy(kp.pk[:], pk)

	return kp, nil
}

func (pair *KeyPair) Public() []byte {
	return pair.pk[:]
}

func (pair *KeyPair) ClientSessionKeys(serverPk []byte) ([]byte, []byte, error) {
	q, err := curve25519.X25519(pair.sk[:], serverPk)
	if err != nil {
		return nil, nil, errors.New("failed to compute share point in x25519:\n\t" + err.Error())
	}

	h, err := blake2b.New(2*config.SESSION_KEY_SIZE, nil)
	if err != nil {
		return nil, nil, errors.New("failed to new blake2b:\n\t" + err.Error())
	}

	for _, b := range [][]byte{q, pair.Public(), serverPk} {
		if _, err := h.Write(b); err != nil {
			return nil, nil, errors.New("failed to write to hash:\n\t" + err.Error())
		}
	}

	keys := h.Sum(nil)

	return keys[:config.SESSION_KEY_SIZE], keys[config.SESSION_KEY_SIZE:], nil
}

func (pair *KeyPair) ServerSessionKeys(clientPk []byte) ([]byte, []byte, error) {
	q, err := curve25519.X25519(pair.sk[:], clientPk)
	if err != nil {
		return nil, nil, errors.New("failed to compute share point in x25519:\n\t" + err.Error())
	}

	h, err := blake2b.New(2*config.SESSION_KEY_SIZE, nil)
	if err != nil {
		return nil, nil, errors.New("failed to new blake2b:\n\t" + err.Error())
	}

	for _, b := range [][]byte{q, clientPk, pair.Public()} {
		if _, err := h.Write(b); err != nil {
			return nil, nil, errors.New("failed to write to hash:\n\t" + err.Error())
		}
	}

	keys := h.Sum(nil)

	return keys[:config.SESSION_KEY_SIZE], keys[config.SESSION_KEY_SIZE:], nil
}
