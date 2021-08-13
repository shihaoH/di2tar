/*
Copyright 2020 The KubeSphere Authors.

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
	"github.com/shihaoH/di2tar/pkg/pull"
	"github.com/spf13/cobra"
	"strings"
)

// pullCmd represents the config command
var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull a docker image",
	RunE: func(cmd *cobra.Command, args []string) error {

		if opt.Image == "" && len(args) == 1 {
			opt.Image = args[0]
		}
		reg := opt.Registry
		repo := "library"
		image := opt.Image
		tag := opt.Tag
		if strings.Contains(image, "/") {
			imgparts := strings.Split(opt.Image, "/")
			// Docker client doesn't seem to consider the first element as a potential registry unless there is a '.' or ':'
			if len(imgparts) > 1 {
				if strings.Contains(imgparts[0], ".") || strings.Contains(imgparts[0], ":") {
					reg = imgparts[0]
					repo = strings.Join(imgparts[1:len(imgparts)-1], "/")
				} else {
					repo = strings.Join(imgparts[0:len(imgparts)-1], "/")
				}
				image = imgparts[len(imgparts)-1]
			}
			// split the image and tag
			if strings.Contains(image, "@") {
				im := strings.Split(image, "@")
				if len(im) != 2 {
					return fmt.Errorf("image or tag exception")
				}
				image = im[0]
				tag = im[1]
			} else if strings.Contains(image, ":") {
				im := strings.Split(image, ":")
				if len(im) != 2 {
					return fmt.Errorf("image or tag exception")
				}
				image = im[0]
				tag = im[1]
			}
		}

		repository := fmt.Sprintf("%s/%s", repo, image)
		err := pull.Pull(reg, repository, tag)
		if err != nil {
			return err
		}
		return err
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
}
