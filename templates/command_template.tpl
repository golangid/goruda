package cmd

import (
	{{ .PackageName }} "github.com/golangid/goruda/{{ .TargetDirectory }}"
	"github.com/golangid/goruda/{{ .TargetDirectory }}/internal/delivery"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	RootCMD = &cobra.Command{
		Use:   "{{ .PackageName }}",
	}

	port = ""

	httpServerCMD = &cobra.Command{
		Use:   "http",
		Short: "Run HTTP Server",
		Long:  "Run HTTP Server for {{ .PackageName }}",
		Run: func(cmd *cobra.Command, args []string) {
			e := echo.New()
			service := domain.ServiceImplementation{}
			delivery.RegisterHTTPPath(e, service)
			if port == "" {
				port = "8080"
			}
			logrus.Infof("starting HTTP server with port %v", port)
			if err := e.Start(":"+port); err != nil {
				logrus.Fatal(err)
			}
		},
	}
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	RootCMD.AddCommand(httpServerCMD)
	httpServerCMD.Flags().StringVarP(&port, "port", "p", "",
    		"Port number for the App. Default is 8080.")
}
