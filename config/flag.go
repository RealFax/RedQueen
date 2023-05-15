package config

import "strconv"

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

type validatorStringValue[T stringValidator] struct{ val *T }

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
	*v.val = val.(T)
	return nil
}

func (v validatorStringValue[T]) String() string { return string(*v.val) }
