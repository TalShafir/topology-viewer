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
	"os"

	"github.com/TalShafir/topology-viewer/pkg/cmd"
	"github.com/TalShafir/topology-viewer/pkg/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/genericiooptions"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	label         string
	allNamespaces bool

	topologyViewerOptions *cmd.TopologyViewerOptions

	configFlags *genericclioptions.ConfigFlags
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "topology-viewer",
	Short: "Shows the topology of the cluster accross domains",
	Long: `This plugin shows how the cluster is spread accross different domains.
A domain is a different values of a node label (e.g different values of 'topology.kubernetes.io/zone').
You can view how the nodes themselves are spread accross the topologies or pods with optional label selector.`,
	Annotations: map[string]string{
		cobra.CommandDisplayNameAnnotation: util.PrefixWithKubectl("topology-viewer"),
	},
	PersistentPreRunE: func(command *cobra.Command, args []string) error {
		clientConfig, err := configFlags.ToRESTConfig()
		if err != nil {
			return err
		}

		client, err := kubernetes.NewForConfig(clientConfig)
		if err != nil {
			return err
		}

		topologyViewerOptions = cmd.NewTopologyViewerOptions(client, genericiooptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}, label, configFlags, allNamespaces)

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.topology-viewer.yaml)")
	rootCmd.PersistentFlags().StringVar(&label, "label", "topology.kubernetes.io/zone", "toplogy label to use")
	rootCmd.PersistentFlags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, `If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even
	if specified with --namespace.`)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	configFlags = genericclioptions.NewConfigFlags(true)
	configFlags.AddFlags(rootCmd.Flags())
	configFlags.AddFlags(rootCmd.PersistentFlags())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.AutomaticEnv() // read in environment variables that match
}
