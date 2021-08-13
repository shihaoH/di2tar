/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

type Options struct {
	Registry string
	Image string
	Tag string
	Username string
	Password string
}

var opt Options

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "di2tar",
	Short: "Docker image and compress it to `{imageName:tag}.tar` ",
	Long: ``,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//Run: func(cmd *cobra.Command, args []string) { },
	//RunE: func(cmd *cobra.Command, args []string) error {
	//	fmt.Println("di2tar")
	//},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	_ = exec.Command("/bin/bash", "-c", "ulimit -u 65535").Run()
	_ = exec.Command("/bin/bash", "-c", "ulimit -n 65535").Run()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVarP(&opt.Image, "image", "i", "", "Image")
	rootCmd.PersistentFlags().StringVarP(&opt.Registry, "registry", "r", "registry-1.docker.io", "Registry")
	rootCmd.PersistentFlags().StringVarP(&opt.Tag, "tag", "t", "latest", "Tag")
	rootCmd.PersistentFlags().StringVarP(&opt.Username, "user", "u", "", "Username")
	rootCmd.PersistentFlags().StringVarP(&opt.Password, "pwd", "p", "", "Password")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("debug", "", true, "Debug to the cmd")
}

