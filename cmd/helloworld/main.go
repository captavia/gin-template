package main

import (
	"gin-template/internal/service"
	"gin-template/pkg/provider"
	"github.com/samber/do"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	root.AddCommand(helloworld)
	execErr := root.Execute()
	if execErr != nil {
		logrus.Fatal(execErr)
	}
}

var root = &cobra.Command{
	Use: "",
}

var helloworld = &cobra.Command{
	Use: "run",
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Info("execute run")

		di := do.New()
		{
			do.Provide(di, provider.Formatter())
			do.Provide(di, provider.Config("./config.toml"))
		}

		app := service.NewService(di)
		if listenErr := app.Run("0.0.0.0:8000"); listenErr != nil {
			logrus.Fatal(listenErr)
		}
	},
}
