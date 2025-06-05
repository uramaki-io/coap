package coap

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func EquateOptions() cmp.Option {
	return cmp.Options{
		cmp.Transformer("Options", func(o Options) []string {
			var opts []string
			for _, opt := range o.data {
				opts = append(opts, opt.String())
			}
			return opts
		}),
		cmpopts.IgnoreUnexported(Options{}),
	}
}
