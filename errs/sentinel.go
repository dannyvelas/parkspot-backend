package errs

// Sentinel is an error that also carries its own stable identity label, so
// code further up a wrap chain (e.g. metrics labeling) can classify it via a
// single errors.As instead of maintaining a separate list of every sentinel.
type Sentinel struct {
	label string
	msg   string
}

func NewSentinel(label, msg string) *Sentinel {
	return &Sentinel{label: label, msg: msg}
}

func (s *Sentinel) Error() string {
	return s.msg
}

func (s *Sentinel) Label() string {
	return s.label
}
