package flags

import "github.com/spf13/pflag"

func MapWithPrefix(f *pflag.FlagSet, name string, errorHandling pflag.ErrorHandling, prefix string) *pflag.FlagSet {
	fNew := pflag.NewFlagSet(name, errorHandling)
	f.VisitAll(func(flag *pflag.Flag) {
		if prefix != "" {
			flag.Name = prefix + "_" + flag.Name
		}
		fNew.AddFlag(flag)
	})

	return fNew
}
