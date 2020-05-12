package gamming

import (
	"github.com/tanema/gween"
)

type Action struct {
	Nexts    []func(tweens *[]gween.Tween)
	OnChange func(float32)
	OnFinish []func()
}

func (a *Action) addOnFinish(f func()) {
	if a.OnFinish == nil {
		a.OnFinish = make([]func(), 0)
	}
	a.OnFinish = append(a.OnFinish, f)
}
