package pocketenv

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/nacl/box"
)

const defaultPublicKey = "2bf96e12d109e6948046a7803ef1696e12c11f04f20a6ce64dbd4bcd93db9341"

// sealedBoxEncrypt implements libsodium's crypto_box_seal using X25519 + XSalsa20-Poly1305.
// The nonce is derived as BLAKE2b-24(epk || rpk), matching libsodium's sealed-box construction.
func sealedBoxEncrypt(recipientPublicKeyHex, message string) (string, error) {
	rpkBytes, err := hex.DecodeString(recipientPublicKeyHex)
	if err != nil {
		return "", fmt.Errorf("pocketenv: invalid public key hex: %w", err)
	}
	if len(rpkBytes) != 32 {
		return "", fmt.Errorf("pocketenv: public key must be 32 bytes, got %d", len(rpkBytes))
	}

	var rpk [32]byte
	copy(rpk[:], rpkBytes)

	// Generate an ephemeral X25519 keypair.
	epk, esk, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return "", fmt.Errorf("pocketenv: failed to generate ephemeral key: %w", err)
	}

	// Nonce = first 24 bytes of BLAKE2b(epk || rpk).
	h, err := blake2b.New(24, nil)
	if err != nil {
		return "", fmt.Errorf("pocketenv: blake2b init: %w", err)
	}
	h.Write(epk[:])
	h.Write(rpk[:])
	nonceSlice := h.Sum(nil)
	var nonce [24]byte
	copy(nonce[:], nonceSlice)

	// Encrypt: output = epk || box.Seal(message, nonce, rpk, esk)
	encrypted := box.Seal(nil, []byte(message), &nonce, &rpk, esk)
	result := make([]byte, 32+len(encrypted))
	copy(result[:32], epk[:])
	copy(result[32:], encrypted)

	return base64.RawURLEncoding.EncodeToString(result), nil
}

// redact masks the middle of a value, keeping the first 11 and last 3 characters visible.
// Values of 14 characters or fewer are returned unchanged.
func redact(value string) string {
	if len(value) <= 14 {
		return value
	}
	return value[:11] + strings.Repeat("*", 24) + value[len(value)-3:]
}

// redactSSHKey masks the body of an OpenSSH private key, preserving the first 10
// and last 5 non-newline characters. Matches the TypeScript sshkeys.ts inline logic.
func redactSSHKey(privateKey string) string {
	const header = "-----BEGIN OPENSSH PRIVATE KEY-----"
	const footer = "-----END OPENSSH PRIVATE KEY-----"

	headerIdx := strings.Index(privateKey, header)
	footerIdx := strings.Index(privateKey, footer)

	if headerIdx == -1 || footerIdx == -1 {
		return strings.ReplaceAll(privateKey, "\n", `\n`)
	}

	body := privateKey[headerIdx+len(header) : footerIdx]
	chars := []rune(body)

	var nonNewlineIdxs []int
	for i, c := range chars {
		if c != '\n' {
			nonNewlineIdxs = append(nonNewlineIdxs, i)
		}
	}

	if len(nonNewlineIdxs) > 15 {
		for _, i := range nonNewlineIdxs[10 : len(nonNewlineIdxs)-5] {
			chars[i] = '*'
		}
	}

	result := header + string(chars) + footer
	return strings.ReplaceAll(result, "\n", `\n`)
}
