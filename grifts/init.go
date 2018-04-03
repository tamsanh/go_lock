package grifts

import (
	"github.com/gobuffalo/buffalo"
	"github.com/tamsanh/go_lock2/actions"
)

func init() {
	buffalo.Grifts(actions.App())
}
