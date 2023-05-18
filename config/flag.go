package config

import (
	"github.com/pkg/errors"
	"net/netip"
	"strconv"
	"strings"
)

const (
	usage = `
Usage of RedQueen:
	
	method: server
	example: 
		./RedQueen [method] <options>
`
	serverUsage = `
Usage of RedQueen(server):

	example: ./RedQueen server -config-file ./config.toml
`
)

// -- uint32 value --

type uin32Value uint32

func newUInt32Value(val uint32, p *uint32) *uin32Value {
	*p = val
	return (*uin32Value)(p)
}

func (u *uin32Value) Set(s string) error {
	val, err := strconv.ParseUint(s, 0, 32)
	if err != nil {
		return err
	}
	*u = uin32Value(val)
	return nil
}

func (u *uin32Value) String() string { return strconv.FormatUint(uint64(*u), 10) }

// -- EnumStoreBackend value --

type enumStoreBackendValue EnumStoreBackend

func newEnumStoreBackendValue(val string, p *EnumStoreBackend) *enumStoreBackendValue {
	*p = EnumStoreBackend(val)
	return (*enumStoreBackendValue)(p)
}

func (e *enumStoreBackendValue) Set(s string) error {
	val := EnumStoreBackend(s)
	if err := val.Valid(); err != nil {
		return err
	}
	*e = enumStoreBackendValue(val)
	return nil
}

func (e *enumStoreBackendValue) String() string { return string(*e) }

// -- stringValidator value --

type validatorStringValue[T stringValidator] struct{ ptr *T }

func newValidatorStringValue[T stringValidator](val string, p *T) *validatorStringValue[T] {
	*p = T(val)
	return &validatorStringValue[T]{p}
}

func (v validatorStringValue[T]) Set(s string) error {
	val := any(T(s))
	if validator, ok := val.(Validator); ok {
		if err := validator.Valid(); err != nil {
			return err
		}
	}
	*v.ptr = val.(T)
	return nil
}

func (v validatorStringValue[T]) String() string {
	if v.ptr == nil {
		return ""
	}
	return string(*v.ptr)
}

// -- []ClusterBootstrap value --

func DecodeClusterBootstraps(s string) ([]ClusterBootstrap, error) {
	if s == "" {
		return []ClusterBootstrap{}, nil
	}

	cs := strings.Split(s, ",")
	if len(cs) == 0 {
		return nil, errors.New("cluster not found")
	}

	cbs := make([]ClusterBootstrap, len(cs))
	for i, c := range cs {
		res := strings.Split(c, "@")
		if len(res) != 2 {
			return nil, errors.Errorf("invalid cluster info: %s", c)
		}

		if _, err := netip.ParseAddrPort(res[1]); err != nil {
			return nil, errors.Wrap(err, "invalid peer addr format")
		}

		cbs[i] = ClusterBootstrap{
			Name:     res[0],
			PeerAddr: res[1],
		}

	}

	return cbs, nil
}

func EncodeClusterBootstraps(s []ClusterBootstrap) string {
	if len(s) == 0 {
		return ""
	}
	builder := strings.Builder{}
	for _, cluster := range s {
		builder.WriteString(cluster.Name)
		builder.WriteRune('@')
		builder.WriteString(cluster.PeerAddr)
		builder.WriteRune(',')
	}
	return builder.String()[:builder.Len()-1]
}

type clusterBootstrapsValue []ClusterBootstrap

func newClusterBootstrapsValue(val string, p *[]ClusterBootstrap) *clusterBootstrapsValue {
	*p, _ = DecodeClusterBootstraps(val)
	return (*clusterBootstrapsValue)(p)
}

func (v *clusterBootstrapsValue) Set(s string) error {
	clusters, err := DecodeClusterBootstraps(s)
	if err != nil {
		return err
	}
	*v = clusters
	return nil
}

func (v *clusterBootstrapsValue) String() string { return EncodeClusterBootstraps(*v) }
