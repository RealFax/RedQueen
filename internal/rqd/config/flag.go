package config

import (
	"net/netip"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	usage = `
Usage of RedQueen:
	
	method: server
	format: 
		./RedQueen [method] <options>
`
	serverUsage = `
Usage of RedQueen(server):

	example: ./RedQueen server -config-file ./config.toml
`
)

var (
	errParse = errors.New("parse error")
)

// -- bool Value
type boolValue bool

func newBoolValue(val bool, p *bool) *boolValue {
	*p = val
	return (*boolValue)(p)
}

func (b *boolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		err = errParse
	}
	*b = boolValue(v)
	return err
}

func (b *boolValue) String() string { return strconv.FormatBool(bool(*b)) }

// -- string Value
type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) String() string { return string(*s) }

// -- int64 Value
type int64Value int64

func newInt64Value(val int64, p *int64) *int64Value {
	*p = val
	return (*int64Value)(p)
}

func (i *int64Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	*i = int64Value(v)
	return err
}

func (i *int64Value) String() string { return strconv.FormatInt(int64(*i), 10) }

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
		return nil, errors.New("invalid cluster bootstraps")
	}

	cs := strings.Split(s, ",")
	if len(cs) == 0 {
		return nil, errors.New("empty cluster bootstraps")
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

// -- map[string]string value --

func decodeStringMap(s string) (map[string]string, error) {
	if s == "" {
		return nil, errors.New("invalid string map")
	}
	entries := strings.Split(s, ",")
	if len(entries) == 0 {
		return nil, errors.New("empty string map")
	}
	m := make(map[string]string)
	for _, entry := range entries {
		kv := strings.Split(entry, ":")
		if len(kv) != 2 {
			continue
		}
		m[kv[0]] = kv[1]
	}
	return m, nil
}

func encodeStringMap(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	builder := strings.Builder{}
	for k, v := range m {
		builder.WriteString(k)
		builder.WriteRune(':')
		builder.WriteString(v)
		builder.WriteRune(',')
	}
	return builder.String()[:builder.Len()-1]
}

type stringMapValue map[string]string

func (v *stringMapValue) Set(s string) error {
	entries, err := decodeStringMap(s)
	if err != nil {
		return err
	}
	*v = entries
	return nil
}

func (v *stringMapValue) String() string { return encodeStringMap(*v) }

func newStringMap(val string, p *map[string]string) *stringMapValue {
	*p, _ = decodeStringMap(val)
	return (*stringMapValue)(p)
}
