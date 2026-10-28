package clipboard

type Clipboard interface {
	Read() ([]byte, error)
	Write(data []byte) error
}

var constructor func() Clipboard

func NewManager() Clipboard {
	if constructor == nil {
		panic("clipboard: constructor not initialized (missing platform implementation)")
	}
	return constructor()
}
