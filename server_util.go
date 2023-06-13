package red

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"github.com/RealFax/RedQueen/config"
	"github.com/RealFax/RedQueen/store"
	"github.com/RealFax/RedQueen/store/nuts"
	"math/big"
	"path/filepath"
	"time"
)

func GenX509CertificateTemplate(host string, pub *ecdsa.PublicKey) (*x509.Certificate, error) {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil, errors.New("invalid ecdsa public key")
	}
	ski := sha256.Sum256(elliptic.Marshal(elliptic.P256(), pub.X, pub.Y))
	now := time.Now()
	return &x509.Certificate{
		SerialNumber: big.NewInt(now.UnixNano()),
		Subject: pkix.Name{
			Country:            []string{"None"},
			Organization:       []string{"RedQueen"},
			OrganizationalUnit: []string{"server"},
			CommonName:         host,
		},
		NotBefore:             now,
		NotAfter:              now.AddDate(1, 0, 0),
		SubjectKeyId:          ski[:],
		BasicConstraintsValid: true,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}, nil
}
func GenX509KeyPair(host string) (*tls.Certificate, error) {
	pri, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	certTemplate, err := GenX509CertificateTemplate(host, &pri.PublicKey)
	if err != nil {
		return nil, err
	}

	cert, err := x509.CreateCertificate(rand.Reader, certTemplate, certTemplate, pri.Public(), pri)
	if err != nil {
		return nil, err
	}

	out := tls.Certificate{}
	out.Certificate = append(out.Certificate, cert)
	out.PrivateKey = pri
	// out.SupportedSignatureAlgorithms = []tls.SignatureScheme{tls.ECDSAWithP256AndSHA256}

	return &out, nil
}

func newNutsStore(cfg config.Store, dir string) (store.Store, error) {
	if cfg.Nuts.StrictMode {
		nuts.EnableStrictMode()
	} else {
		nuts.DisableStrictMode()
	}

	return nuts.New(nuts.Config{
		NodeNum: cfg.Nuts.NodeNum,
		Sync:    cfg.Nuts.Sync,
		DataDir: filepath.Join(dir, "fsm"),
	})
}

func newStoreBackend(cfg config.Store, dir string) (store.Store, error) {
	handle, ok := map[config.EnumStoreBackend]func(config.Store, string) (store.Store, error){
		config.StoreBackendNuts: newNutsStore,
	}[cfg.Backend]
	if !ok {
		return nil, errors.New("unsupported store backend")
	}
	return handle(cfg, dir)
}

func loadTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates:     []tls.Certificate{cert},
		MinVersion:       tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}, nil
}

func generateTLSConfig(host string) (*tls.Config, error) {
	cert, err := GenX509KeyPair(host)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{*cert},
	}, nil
}
