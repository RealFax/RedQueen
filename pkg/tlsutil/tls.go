package tlsutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"time"
)

func GenX509KeyPairSubjectKeyID() []byte {
	p := make([]byte, 10)
	_, _ = rand.Read(p)
	return p
}

func GenX509KeyPair(subject pkix.Name) (tls.Certificate, error) {
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(now.Unix()),
		Subject:               subject,
		NotBefore:             now,
		NotAfter:              now.AddDate(10, 0, 0),
		SubjectKeyId:          GenX509KeyPairSubjectKeyID(),
		BasicConstraintsValid: true,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	cert, err := x509.CreateCertificate(rand.Reader, template, template, priv.Public(), priv)
	if err != nil {
		return tls.Certificate{}, nil
	}

	var outCert tls.Certificate
	outCert.Certificate = append(outCert.Certificate, cert)
	outCert.PrivateKey = priv

	return outCert, nil
}
