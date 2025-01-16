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

var labelSelector string

// podCmd represents the pod command
var podCmd = &cobra.Command{
	Use:     "pod",
	Aliases: []string{"pods"},
	Short:   "Shows how the pods are spread accross topologies",
	Long: `Shows how pods are spread accross topologies, including count and any other allocatable resources of the pods.
'-' means that the resource wasn't present on any of the pods in the domain.
If no label selector was provided all pods will be shown.
A pod is considered to be part of a domain according to the value of the label on the node its running on.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		toplogies, err := topologyViewerOptions.Pods(cmd.Context(), labelSelector)
		if err != nil {
			return err
		}

		util.PrintTopologies(toplogies, topologyViewerOptions.Out, includeMembers)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(podCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// podCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// podCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	podCmd.Flags().StringVarP(&labelSelector, "selector", "l", "", `Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2). Matching
        pods must satisfy all of the specified label constraints.`)
}
