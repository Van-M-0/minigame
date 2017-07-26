package network

type netContext struct {
	values 		map[string]interface{}
}

func (nt *netContext) Set(key string, val interface{}) {
	nt.values[key] = val
}

func (nt *netContext) Get(key string) interface{} {
	if val, ok := nt.values[key]; !ok {
		return nil
	} else {
		return val
	}
}

func newNetContext() *netContext {
	return &netContext{
		values: make(map[string]interface{}),
	}
}

