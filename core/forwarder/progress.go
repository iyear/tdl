package forwarder

type ProgressClone interface {
	OnClone(elem Elem, state ProgressState)
}

type Progress interface {
	OnAdd(elem Elem)
	ProgressClone
	OnDone(elem Elem, err error)
}

type ProgressState struct {
	Done  int64
	Total int64
}
