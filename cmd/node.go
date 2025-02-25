/*
Copyright © 2024 Tal Shafir

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
	"github.com/TalShafir/topology-viewer/pkg/util"
	"github.com/spf13/cobra"
)

// nodeCmd represents the node command
var nodeCmd = &cobra.Command{
	Use:     "node",
	Aliases: []string{"nodes"},
	Short:   "Shows how the nodes are spread accross topologies",
	Long: `Shows how the nodes spread accross topologies, including count and any other allocatable resources of the nodes.
The output includes the node count and all allocatable resources of the nodes
'-' means that the resource wasn't present on any of the nodes in the domain`,
	RunE: func(cmd *cobra.Command, args []string) error {
		toplogies, err := topologyViewerOptions.Nodes(cmd.Context())
		if err != nil {
			return err
		}

		util.PrintTopologies(toplogies, topologyViewerOptions.Out, includeMembers)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(nodeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// nodeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// nodeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
