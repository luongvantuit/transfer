package cipher

import (
	"crypto/sha256"
	"math/rand"
	"strings"
)

// add to struct:
type SubstitutionCipher struct {
	enc [256]byte
	dec [256]byte

	// special mapping for first digit: domain { '1'..'9' } -> { '1'..'9' }
	firstDigitEnc [10]byte // use index 1..9
	firstDigitDec [10]byte // use index 1..9
}

// update NewSubstitutionCipher: create 2 number permutations
func NewSubstitutionCipher(key string) Cipher {
	var c SubstitutionCipher

	// identity
	for i := 0; i < 256; i++ {
		c.enc[i] = byte(i)
		c.dec[i] = byte(i)
	}

	seed := seedFromKey(key)
	r := rand.New(rand.NewSource(seed))

	up := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	lo := []byte("abcdefghijklmnopqrstuvwxyz")
	dg := []byte("0123456789")
	dgFirst := []byte("123456789") // only for first digit

	upSh := append([]byte(nil), up...)
	loSh := append([]byte(nil), lo...)
	dgSh := append([]byte(nil), dg...)
	dgFirstSh := append([]byte(nil), dgFirst...)

	r.Shuffle(len(upSh), func(i, j int) { upSh[i], upSh[j] = upSh[j], upSh[i] })
	r.Shuffle(len(loSh), func(i, j int) { loSh[i], loSh[j] = loSh[j], loSh[i] })
	r.Shuffle(len(dgSh), func(i, j int) { dgSh[i], dgSh[j] = dgSh[j], dgSh[i] })
	r.Shuffle(len(dgFirstSh), func(i, j int) { dgFirstSh[i], dgFirstSh[j] = dgFirstSh[j], dgFirstSh[i] })

	// letters (keep as before)
	for i := range up {
		c.enc[up[i]] = upSh[i]
		c.dec[upSh[i]] = up[i]
	}
	for i := range lo {
		c.enc[lo[i]] = loSh[i]
		c.dec[loSh[i]] = lo[i]
	}
	// numbers for non-first positions: domain 0..9
	for i := range dg {
		c.enc[dg[i]] = dgSh[i]
		c.dec[dgSh[i]] = dg[i]
	}
	// numbers for first position: domain 1..9
	for i := range dgFirst { // i = 0..8 represents '1'..'9'
		d := dgFirst[i]   // '1'..'9'
		m := dgFirstSh[i] // map to '1'..'9'
		c.firstDigitEnc[d-'0'] = m
		c.firstDigitDec[m-'0'] = d
	}
	return &c
}

// seedFromKey generates a seed from the given key using SHA-256
func seedFromKey(key string) int64 {
	sum := sha256.Sum256([]byte(key))
	var seed int64
	for i := 0; i < 8; i++ {
		seed = (seed << 8) | int64(sum[i])
	}
	return seed
}

// Encrypt/Decrypt: ASCII-only (byte-wise). Non-ASCII characters are kept as bytes.
func (c *SubstitutionCipher) Encrypt(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		ch := s[i]
		b.WriteByte(c.enc[ch])
	}
	return b.String()
}

func (c *SubstitutionCipher) Decrypt(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		ch := s[i]
		b.WriteByte(c.dec[ch])
	}
	return b.String()
}

// EncryptNumber encrypts a (possibly signed) numeric string.
// Behavior:
//   - Optional leading '-' is preserved.
//   - The first digit of the numeric part is mapped via the {1..9} table,
//     ensuring the output never starts with '0'.
//   - If the numeric part starts with '0', we map it with the normal table;
//     if it still becomes '0', we replace it with a deterministic non-zero
//     (mapping of '1' from the first-digit table).
//   - Non-digit characters (besides an optional leading '-') cause a no-op (returns s).
//   - "-0" is normalized to "0".
func (c *SubstitutionCipher) EncryptNumber(s string) string {
	if len(s) == 0 {
		return s
	}

	// handle optional leading '-'
	neg := false
	num := s
	if s[0] == '-' {
		neg = true
		num = s[1:]
		if len(num) == 0 {
			return s // "-" only -> no-op
		}
	}

	// validate digits
	for i := 0; i < len(num); i++ {
		if num[i] < '0' || num[i] > '9' {
			return s // policy: non-digit -> no-op
		}
	}

	// normalize "-0" -> "0"
	if neg && num == "0" {
		return "0"
	}

	// empty numeric (shouldn't happen here), just return
	if len(num) == 0 {
		return s
	}

	out := make([]byte, len(num))

	// first digit (ensure non-zero output)
	if num[0] < '1' || num[0] > '9' {
		// input starts with '0'
		m := c.enc[num[0]]
		if m == '0' {
			// deterministic non-zero fallback
			m = c.firstDigitEnc[1]
		}
		out[0] = m
	} else {
		out[0] = c.firstDigitEnc[num[0]-'0']
	}

	// remaining digits
	for i := 1; i < len(num); i++ {
		out[i] = c.enc[num[i]]
	}

	if neg {
		return "-" + string(out)
	}
	return string(out)
}

// DecryptNumber decrypts a (possibly signed) numeric string produced by EncryptNumber.
// Behavior:
//   - Preserves an optional leading '-'.
//   - First digit of the numeric part is reversed via the {1..9} reverse table
//     if it is in '1'..'9'; otherwise falls back to the normal table.
//   - Non-digit characters (besides an optional leading '-') cause a no-op (returns s).
//   - "-0" is normalized back to "0".
func (c *SubstitutionCipher) DecryptNumber(s string) string {
	if len(s) == 0 {
		return s
	}

	// handle optional leading '-'
	neg := false
	num := s
	if s[0] == '-' {
		neg = true
		num = s[1:]
		if len(num) == 0 {
			return s // "-" only -> no-op
		}
	}

	// validate digits
	for i := 0; i < len(num); i++ {
		if num[i] < '0' || num[i] > '9' {
			return s // policy: non-digit -> no-op
		}
	}

	// decrypt numeric part
	out := make([]byte, len(num))

	// first digit
	if num[0] >= '1' && num[0] <= '9' {
		out[0] = c.firstDigitDec[num[0]-'0']
	} else {
		out[0] = c.dec[num[0]]
	}

	// remaining digits
	for i := 1; i < len(num); i++ {
		out[i] = c.dec[num[i]]
	}

	plain := string(out)

	// normalize "-0" -> "0"
	if plain == "0" {
		return "0"
	}
	if neg {
		return "-" + plain
	}
	return plain
}
