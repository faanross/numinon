package listener

// commonListener holds the fields and implements the helper methods
// common to all listener types. It is embedded in concrete
// listener implementations.
type commonListener struct {
	id     string
	addr   string
	config ListenerConfig
}

// Addr returns the listener's network address.
func (l *commonListener) Addr() string {
	return l.addr
}

// ID returns the listener's unique identifier.
func (l *commonListener) ID() string {
	return l.id
}

// Type returns the type of the listener.
func (l *commonListener) Type() ListenerType {
	return l.config.Type
}
