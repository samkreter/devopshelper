package cmd


import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/samkreter/VSTSAutoReviewer/conf"
)

func run(cmd *cobra.Command, args []string) {
	err := conf.LoadConfig(cmd)
	if err != nil {
	  log.Fatal("Failed to load config: " + err.Error())
	}
  
	fmt.Printf("+%v\n", config)
  }

// RootCommand will setup and return the root command
func RootCommand() *cobra.Command {
	rootCmd := cobra.Command{
	  Use: "example",
	  Run: run,
	}
  
	// this is where we will configure everything!
	rootCmd.Flags().IntP("port", "p", 0, "the port to do things on")
  
	return &rootCmd
  }
  
  func run(cmd *cobra.Command, args []string) {
	fmt.Println("--- here ---")
  }