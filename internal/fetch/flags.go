package fetch

import "flag"

type pprofFlagSet struct {
	*flag.FlagSet
	args []string
}

func (f *pprofFlagSet) StringList(o, d, c string) *[]*string {
	return &[]*string{f.String(o, d, c)}
}

func (f *pprofFlagSet) ExtraUsage() string {
	return ""
}

func (f *pprofFlagSet) Parse(usage func()) []string {
	f.Usage = usage
	f.FlagSet.Parse(f.args)
	return f.Args()
}
