package socks5

type DecryptState int
type EncryptState int

const (
	DecryptSuccess = iota
	DecryptFailed
	CommonConnection
)

type handshakeEncryptFunc func([]byte) []byte

func versionHandshakeRequestEncrypt(origin []byte) ([]byte, EncryptState) {
	return origin, DecryptFailed
}

func versionHandshakeRequestDecrypt(origin []byte) []byte {
	return origin
}
