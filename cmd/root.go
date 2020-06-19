package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	RootCMD = &cobra.Command{
		Use:   "goruda",
		Short: "API generator from openapi documentation",
		Long:  "See https://github.com/golangid/goruda for more information",
	}
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	generatorCMD.Flags().StringVarP(&packageName, "package", "p", "",
		"Package name for the generated file. Default is domain")
	generatorCMD.Flags().StringVarP(&pathToGeneratedFiles, "out", "o", "",
		"Path to target directory. Default is ./generated")

	RootCMD.AddCommand(generatorCMD)
}
