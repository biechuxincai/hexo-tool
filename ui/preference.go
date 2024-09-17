package ui

import (
	"reflect"
	"sync"
)

type DiskPreferences struct {
	path            string
	values          map[string]any
	lock            sync.RWMutex
	changeListeners []func()
}

func (p *DiskPreferences) SetString(key string, value string) {
	//p.set(key, value)
}

func (p *DiskPreferences) get(key string) (any, bool) {
	p.lock.RLock()
	defer p.lock.RUnlock()

	v, err := p.values[key]
	return v, err
}

func (p *DiskPreferences) remove(key string) {
	p.lock.Lock()
	delete(p.values, key)
	p.lock.Unlock()

	//p.fireChange()
}

func (p *DiskPreferences) set(key string, value any) {
	p.lock.Lock()

	if reflect.TypeOf(value).Kind() == reflect.Slice {
		s := reflect.ValueOf(value)
		old := reflect.ValueOf(p.values[key])
		if p.values[key] != nil && s.Len() == old.Len() {
			changed := false
			for i := 0; i < s.Len(); i++ {
				if s.Index(i).Interface() != old.Index(i).Interface() {
					changed = true
					break
				}
			}
			if !changed {
				p.lock.Unlock()
				return
			}
		}
	} else {
		if stored, ok := p.values[key]; ok && stored == value {
			p.lock.Unlock()
			return
		}
	}

	p.values[key] = value
	p.lock.Unlock()

	//p.fireChange()
}
