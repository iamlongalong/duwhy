/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"context"
	"duwhy/core/memprovider"
	"duwhy/internal/server"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var (
	DuLogFile  *string
	DuIgnores  *[]string
	ServerHost *string
	ServerPort *int

	ServerAuthUser *string
	ServerAuthPass *string
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "run http server",
	Long:  `run http server for disklog`,
	Run: func(cmd *cobra.Command, args []string) {
		pb, err := memprovider.NewMemDuFileBuilder(*DuLogFile, &memprovider.MemDUBuilderOption{
			Ignore: *DuIgnores,
		})
		if err != nil {
			log.Printf("run duserver fail : %s\n", err)
			os.Exit(1)
			return
		}

		p, err := pb.Build()
		if err != nil {
			log.Printf("run duserver fail : %s\n", err)
			os.Exit(1)
			return
		}

		cfg := server.ServerConfig{}
		if *ServerAuthUser != "" {
			cfg.Auth.Enable = true
			cfg.Auth.UserName = *ServerAuthUser
			cfg.Auth.Password = *ServerAuthPass
		}

		cfg.Host = *ServerHost
		cfg.Port = *ServerPort

		server.Serve(context.Background(), server.ServerOption{
			IProvider: p,
			Server:    cfg,
		})
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	ServerHost = serverCmd.Flags().String("host", "0.0.0.0", "server listen host")
	ServerPort = serverCmd.Flags().IntP("port", "p", 8080, "server listen port")
	ServerAuthUser = serverCmd.Flags().StringP("auth.user", "U", "", "server auth user")
	ServerAuthPass = serverCmd.Flags().StringP("auth.pass", "P", "", "server auth pass")

	DuLogFile = serverCmd.Flags().StringP("dufile", "f", "", "du file path")

	DuIgnores = serverCmd.Flags().StringArrayP("ignores", "i", []string{}, "ignore du file paths, like ./xxx/*")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
