package cipher

import (
	"errors"
	"strings"

	"github.com/capitalone/fpe/ff1"
)

// ---------------------------
// helpers: char classes
// ---------------------------
func isDigit(b byte) bool { return '0' <= b && b <= '9' }
func isUpper(b byte) bool { return 'A' <= b && b <= 'Z' }
func isLower(b byte) bool { return 'a' <= b && b <= 'z' }

// digits used by ff1 (0..35 -> '0'..'9','A'..'Z')
var digits36 = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

// map 0..25 -> single-char digit for radix 26 (0..9,A..P)
func encRadix26(v int) byte { return digits36[v] } // v in [0,25]

// map digit char ('0'..'9','A'..'P') -> 0..25
func decRadix26(ch byte) (int, error) {
	if ch >= '0' && ch <= '9' {
		return int(ch - '0'), nil
	}
	if ch >= 'A' && ch <= 'P' {
		return 10 + int(ch-'A'), nil
	}
	return 0, errors.New("invalid radix26 digit")
}

// ---------------------------
// FPE cipher (FF1) per class
// ---------------------------
type FPECipher struct {
	ffDigits *ff1.Cipher // radix 10
	ffUpper  *ff1.Cipher // radix 26 (A..Z)
	ffLower  *ff1.Cipher // radix 26 (a..z)
}

// NewFPECipher builds FF1 ciphers for digits/upper/lower.
// key must be 16, 24, or 32 bytes (AES-128/192/256).
func NewFPECipher(key []byte) (*FPECipher, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, errors.New("key length must be 16, 24, or 32 bytes")
	}
	// NIST FF1 allows a tweak (like a nonce/salt). Keep short & constant per domain.
	tweakD := []byte("D-TWEAK") // <= 8 bytes in this implementation
	tweakU := []byte("U-TWEAK")
	tweakL := []byte("L-TWEAK")

	cD, err := ff1.NewCipher(10, 8, key, tweakD)
	if err != nil {
		return nil, err
	}
	cU, err := ff1.NewCipher(26, 8, key, tweakU)
	if err != nil {
		return nil, err
	}
	cL, err := ff1.NewCipher(26, 8, key, tweakL)
	if err != nil {
		return nil, err
	}
	return &FPECipher{ffDigits: &cD, ffUpper: &cU, ffLower: &cL}, nil
}

// ---------------------------
// encryption / decryption
// ---------------------------

// EncryptPreserving:
// - digits runs -> FF1 (radix10) with cycle-walking to avoid leading '0'
// - uppercase runs -> FF1 (radix26) after mapping A..Z <-> 0..25
// - lowercase runs -> FF1 (radix26) after mapping a..z <-> 0..25
// - '-' before a digit run is preserved (signed numbers)
// - other bytes are copied as-is
func (c *FPECipher) EncryptPreserving(s string) (string, error) {
	var out strings.Builder
	for i := 0; i < len(s); {
		b := s[i]

		// signed number: keep '-' and encrypt following digits as one run
		if b == '-' && i+1 < len(s) && isDigit(s[i+1]) {
			out.WriteByte('-')
			j := i + 1
			for j < len(s) && isDigit(s[j]) {
				j++
			}
			ct, err := c.encryptDigitsNoLeadingZero(s[i+1 : j])
			if err != nil {
				return "", err
			}
			out.WriteString(ct)
			i = j
			continue
		}

		// digit run
		if isDigit(b) {
			j := i
			for j < len(s) && isDigit(s[j]) {
				j++
			}
			ct, err := c.encryptDigitsNoLeadingZero(s[i:j])
			if err != nil {
				return "", err
			}
			out.WriteString(ct)
			i = j
			continue
		}

		// uppercase run
		if isUpper(b) {
			j := i
			for j < len(s) && isUpper(s[j]) {
				j++
			}
			ct, err := c.encryptLetters(s[i:j], true)
			if err != nil {
				return "", err
			}
			out.WriteString(ct)
			i = j
			continue
		}

		// lowercase run
		if isLower(b) {
			j := i
			for j < len(s) && isLower(s[j]) {
				j++
			}
			ct, err := c.encryptLetters(s[i:j], false)
			if err != nil {
				return "", err
			}
			out.WriteString(ct)
			i = j
			continue
		}

		// passthrough
		out.WriteByte(b)
		i++
	}
	return out.String(), nil
}

// DecryptPreserving is the inverse of EncryptPreserving (same segmentation).
func (c *FPECipher) DecryptPreserving(s string) (string, error) {
	var out strings.Builder
	for i := 0; i < len(s); {
		b := s[i]

		if b == '-' && i+1 < len(s) && isDigit(s[i+1]) {
			out.WriteByte('-')
			j := i + 1
			for j < len(s) && isDigit(s[j]) {
				j++
			}
			pt, err := c.decryptDigitsNoLeadingZero(s[i+1 : j])
			if err != nil {
				return "", err
			}
			out.WriteString(pt)
			i = j
			continue
		}

		if isDigit(b) {
			j := i
			for j < len(s) && isDigit(s[j]) {
				j++
			}
			pt, err := c.decryptDigitsNoLeadingZero(s[i:j])
			if err != nil {
				return "", err
			}
			out.WriteString(pt)
			i = j
			continue
		}

		if isUpper(b) {
			j := i
			for j < len(s) && isUpper(s[j]) {
				j++
			}
			pt, err := c.decryptLetters(s[i:j], true)
			if err != nil {
				return "", err
			}
			out.WriteString(pt)
			i = j
			continue
		}

		if isLower(b) {
			j := i
			for j < len(s) && isLower(s[j]) {
				j++
			}
			pt, err := c.decryptLetters(s[i:j], false)
			if err != nil {
				return "", err
			}
			out.WriteString(pt)
			i = j
			continue
		}

		out.WriteByte(b)
		i++
	}
	return out.String(), nil
}

// ---- digits: FF1 radix10 + cycle-walking to avoid leading '0' ----

func (c *FPECipher) encryptDigitsNoLeadingZero(num string) (string, error) {
	if len(num) == 0 {
		return num, nil
	}
	X := num
	for {
		ct, err := c.ffDigits.Encrypt(X)
		if err != nil {
			return "", err
		}
		if ct[0] != '0' { // accept only if first char is not '0'
			return ct, nil
		}
		// cycle-walk: re-encrypt the ciphertext until constraint satisfied
		X = ct
	}
}

func (c *FPECipher) decryptDigitsNoLeadingZero(ct string) (string, error) {
	if len(ct) == 0 {
		return ct, nil
	}
	X := ct
	for {
		pt, err := c.ffDigits.Decrypt(X)
		if err != nil {
			return "", err
		}
		if pt[0] != '0' { // inverse of the same cycle-walk constraint
			return pt, nil
		}
		X = pt
	}
}

// ---- letters: map to radix26, run FF1, then map back ----

func (c *FPECipher) encryptLetters(seg string, upper bool) (string, error) {
	// map letters -> radix26 digits
	buf := make([]byte, len(seg))
	for i := 0; i < len(seg); i++ {
		var v int
		if upper {
			v = int(seg[i] - 'A') // 0..25
		} else {
			v = int(seg[i] - 'a') // 0..25
		}
		buf[i] = encRadix26(v) // -> '0'..'9','A'..'P'
	}
	X := string(buf)

	var ct string
	var err error
	if upper {
		ct, err = c.ffUpper.Encrypt(X) // still a string of radix26 digits
	} else {
		ct, err = c.ffLower.Encrypt(X)
	}
	if err != nil {
		return "", err
	}

	// map radix26 digits -> letters
	for i := 0; i < len(ct); i++ {
		v, err := decRadix26(ct[i])
		if err != nil {
			return "", err
		}
		if upper {
			buf[i] = byte('A' + v)
		} else {
			buf[i] = byte('a' + v)
		}
	}
	return string(buf), nil
}

func (c *FPECipher) decryptLetters(seg string, upper bool) (string, error) {
	// map letters -> radix26 digits
	buf := make([]byte, len(seg))
	for i := 0; i < len(seg); i++ {
		if upper {
			buf[i] = encRadix26(int(seg[i] - 'A'))
		} else {
			buf[i] = encRadix26(int(seg[i] - 'a'))
		}
	}
	X := string(buf)

	var pt string
	var err error
	if upper {
		pt, err = c.ffUpper.Decrypt(X)
	} else {
		pt, err = c.ffLower.Decrypt(X)
	}
	if err != nil {
		return "", err
	}

	// map radix26 digits -> letters
	for i := 0; i < len(pt); i++ {
		v, err := decRadix26(pt[i])
		if err != nil {
			return "", err
		}
		if upper {
			buf[i] = byte('A' + v)
		} else {
			buf[i] = byte('a' + v)
		}
	}
	return string(buf), nil
}
