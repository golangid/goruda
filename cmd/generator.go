package cmd

import (
	"github.com/golangid/goruda"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	generatorCMD = &cobra.Command{
		Use:   "generate",
		Short: "Generate API based on inserted openapi argument.",
		Long:  "Use generate xxx.yaml to generate the API package",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				logrus.Fatal("missing documentation")
			}
			logrus.Debugf("processed documentation %v", args)

			for _, arg := range args {
				if err := goruda.Generate(arg); err != nil {
					logrus.Fatalf("error generate API: %v", err)
				}
			}
		},
	}
)
