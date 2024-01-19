package red

import (
	"github.com/pkg/errors"

	"github.com/RealFax/RedQueen/api/serverpb"
	"github.com/RealFax/RedQueen/store"
)

type FSMHandlers struct {
	store store.Store
}

func (h *FSMHandlers) swap(namespace *string) (store.Actions, error) {
	if namespace == nil {
		return h.store, nil
	}
	actions, err := h.store.Swap(*namespace)
	if err != nil {
		return nil, errors.Wrap(err, "swap with error")
	}
	return actions, nil
}

func (h *FSMHandlers) SetWithTTL(payload *serverpb.RaftLogPayload) error {
	if payload.Ttl == nil || payload.Key == nil || payload.Value == nil {
		return errors.New("invalid SetWithTTl args")
	}

	dest, err := h.swap(payload.Namespace)
	if err != nil {
		return err
	}

	return dest.SetWithTTL(payload.Key, payload.Value, *payload.Ttl)
}

func (h *FSMHandlers) TrySetWithTTL(payload *serverpb.RaftLogPayload) error {
	if payload.Ttl == nil || payload.Key == nil || payload.Value == nil {
		return errors.New("invalid TrySetWithTTL args")
	}

	dest, err := h.swap(payload.Namespace)
	if err != nil {
		return err
	}

	return dest.TrySetWithTTL(payload.Key, payload.Value, *payload.Ttl)
}

func (h *FSMHandlers) Set(payload *serverpb.RaftLogPayload) error {
	if payload.Key == nil || payload.Value == nil {
		return errors.New("invalid Set args")
	}

	dest, err := h.swap(payload.Namespace)
	if err != nil {
		return err
	}

	return dest.Set(payload.Key, payload.Value)
}

func (h *FSMHandlers) TrySet(payload *serverpb.RaftLogPayload) error {
	if payload.Key == nil || payload.Value == nil {
		return errors.New("invalid TrySet args")
	}

	dest, err := h.swap(payload.Namespace)
	if err != nil {
		return err
	}

	return dest.TrySet(payload.Key, payload.Value)
}

func (h *FSMHandlers) Del(payload *serverpb.RaftLogPayload) error {
	if payload.Key == nil {
		return errors.New("invalid Del args")
	}

	dest, err := h.swap(payload.Namespace)
	if err != nil {
		return err
	}

	return dest.Del(payload.Key)
}

func NewFSMHandlers(s store.Store) map[serverpb.RaftLogCommand]FSMHandleFunc {
	handlers := &FSMHandlers{store: s}

	return map[serverpb.RaftLogCommand]FSMHandleFunc{
		serverpb.RaftLogCommand_SetWithTTL:    handlers.SetWithTTL,
		serverpb.RaftLogCommand_TrySetWithTTL: handlers.TrySetWithTTL,
		serverpb.RaftLogCommand_Set:           handlers.Set,
		serverpb.RaftLogCommand_TrySet:        handlers.TrySet,
		serverpb.RaftLogCommand_Del:           handlers.Del,
	}
}
