package cmd

import (
	"log"

	"gitlab.com/silenteer/go-nats/_example/companyservice/internal/app"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start server ",
	Long:  "Start server ", // add example here
	Run: func(cmd *cobra.Command, args []string) {
		var config app.Config
		err := viper.Unmarshal(&config)
		if err != nil {
			log.Fatalf("unable to decode yaml config to struct, %v", err)
		}
		app.NewServer(&config).Start()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
