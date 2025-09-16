package www

import "embed"

//go:embed *.html
//go:embed *.js
var static embed.FS

func New() *embed.FS {
	return &static
}