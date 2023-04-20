package cosign

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/sigstore/cosign/pkg/cosign"
	"github.com/sigstore/sigstore/pkg/signature"
)

const DockerImagePubKey = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE/IrzBQYeMwvKa44/DF/HB7XDpnE+
f+mU9F/Qbfq25bBWV2+NlYMJv3KvKHNtu3Jknt6yizZjUV4b8WGfKBzFYw==
-----END PUBLIC KEY-----`

func VerifyImageSignature(ctx context.Context, imageRef, pubKey string) (bool, error) {

	// Decode the PEM-encoded public key data
	block, _ := pem.Decode([]byte(pubKey))
	if block == nil {
		return false, fmt.Errorf("failed to decode public key")
	}
	// Parse the public key from the decoded PEM block
	// x509.ParsePKIXPublicKey is used for parsing PKIX public keys, which include RSA, DSA, and ECDSA public keys
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("failed to parse public key: %w", err)
	}
	// Get ecdsa.PublickKey by type assertion
	ecdsaPubKey, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		fmt.Println("public key is not of type *ecdsa.PublicKey")
		panic("fuck!")
	}

	// Get image reference
	ref, err := name.ParseReference(imageRef)
	if err != nil {
		return false, fmt.Errorf("failed to parse image reference: %w", err)
	}

	// Create a signature verifier for an ECDSA signature algorithm using a public key
	// and the SHA256 cryptographic hash function, and then setting the signature verifier as an option for verifying a signed image.
	verifier, err := signature.LoadECDSAVerifier(ecdsaPubKey, crypto.SHA256)
	co := &cosign.CheckOpts{
		SigVerifier: verifier,
	}

	signatures, verified, err := cosign.VerifyImageSignatures(ctx, ref, co)
	for _, sig := range signatures {
		fmt.Fprintf(os.Stdout, "Signature: %s\n", func() string { sig, _ := sig.Base64Signature(); return sig }())
		fmt.Fprintf(os.Stdout, "Payload: %s\n", func() string { payload, _ := sig.Payload(); return string(payload) }())
		fmt.Fprintln(os.Stdout, "====")
	} // Maybe I will use it in future

	return verified, nil

}
