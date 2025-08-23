package cipher

type Cipher interface {
	Encrypt(text string) string
	Decrypt(text string) string
	EncryptNumber(text string) string
	DecryptNumber(text string) string
}
