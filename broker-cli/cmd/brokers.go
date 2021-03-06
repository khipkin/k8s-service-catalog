// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"log"

	"github.com/GoogleCloudPlatform/k8s-service-catalog/broker-cli/client/adapter"
	"github.com/GoogleCloudPlatform/k8s-service-catalog/broker-cli/cmd/flags"
	"github.com/spf13/cobra"
)

var (
	brokersFlags struct {
		host    string
		broker  string
		project string
		title   string
		verbose bool
	}

	// brokersCmd represents the brokers command.
	brokersCmd = &cobra.Command{
		Use:   "brokers",
		Short: "Manage service brokers",
		Long:  "Manage service brokers",
	}

	// brokersCreateCmd represents the brokers create command.
	brokersCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a service broker",
		Long:  "Create a service broker",
		Run: func(cmd *cobra.Command, args []string) {
			flags.CheckFlags(&brokersFlags.project, &brokersFlags.broker)

			// Title defaults to name if not present.
			title := brokersFlags.title
			if title == "" {
				title = brokersFlags.broker
			}

			http := httpAdapterFromFlag()
			res, err := http.CreateBroker(&adapter.CreateBrokerParams{
				Host:    brokersFlags.host,
				Project: brokersFlags.project,
				Name:    brokersFlags.broker,
				Title:   title,
			})
			if err != nil {
				log.Fatalf("Failed to create broker %q in project %q: %v\n", brokersFlags.broker, brokersFlags.project, err)
			}

			fmt.Printf("Successfully created broker %q in project %q!!\n", brokersFlags.broker, brokersFlags.project)
			fmt.Printf("   Title: %s\n", res.Title)
			fmt.Printf("   URL: %s\n", *res.URL)
			fmt.Printf("   Create time: %s\n", *res.CreateTime)
		},
	}

	// brokersDeleteCmd represents the brokers delete command.
	brokersDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a service broker",
		Long:  "Delete a service broker",
		Run: func(cmd *cobra.Command, args []string) {
			flags.CheckFlags(&brokersFlags.project, &brokersFlags.broker)

			http := httpAdapterFromFlag()
			params := &adapter.DeleteBrokerParams{
				BrokerURL: flags.ConstructBrokerURL(brokersFlags.host, brokersFlags.project, brokersFlags.broker),
			}
			err := http.DeleteBroker(params)
			if err != nil {
				log.Fatalf("Failed to delete broker %q in project %q: %v\n", brokersFlags.broker, brokersFlags.project, err)
			}

			fmt.Printf("Successfully deleted broker %q in project %q!!\n", brokersFlags.broker, brokersFlags.project)

		},
	}

	// brokersCleanupCmd represents the command which deletes all service instances and bindings within a broker
	brokersCleanupCmd = &cobra.Command{
		Use:   "cleanup",
		Short: "Delete all service instances and bindings within a broker",
		Long:  "Delete all service instances and bindings within a broker",
		Run: func(cmd *cobra.Command, args []string) {
			flags.CheckFlags(&brokersFlags.project, &brokersFlags.broker)
			brokerURL := flags.ConstructBrokerURL(brokersFlags.host, brokersFlags.project, brokersFlags.broker)
			err := cleanupBroker(brokerURL)
			if err != nil {
				if lir, err := listInstances(brokerURL); err == nil {
					log.Println("The below resources are yet to be cleaned up!!")
					printListInstances(lir)
				}
			} else {
				fmt.Printf("Successfully cleaned up broker %q in project %q!!\n", brokersFlags.broker, brokersFlags.project)
			}
		},
	}

	// brokersListCmd represents the brokers list command.
	brokersListCmd = &cobra.Command{
		Use:   "list",
		Short: "List service brokers in a project",
		Long:  "List service brokers in a project",
		Run: func(cmd *cobra.Command, args []string) {
			flags.CheckFlags(&brokersFlags.project)

			http := httpAdapterFromFlag()
			res, err := http.ListBrokers(&adapter.ListBrokersParams{
				Host:    brokersFlags.host,
				Project: brokersFlags.project})
			if err != nil {
				log.Fatalf("Failed to list brokers in project %q: %v\n", brokersFlags.project, err)
			}

			if len(res.Brokers) == 0 {
				fmt.Printf("Project %q has no associated brokers\n", brokersFlags.project)
				return
			}

			fmt.Printf("Successfully listed brokers in project %q!!\n\n", brokersFlags.project)
			printListBrokers(res)
		},
	}
)

func init() {
	// Flags for all brokers subcommands.
	flags.StringFlag(brokersCmd.PersistentFlags(), &brokersFlags.project, flags.ProjectLongName, flags.ProjectShortName, "[Required] The GCP Project to use.")
	// This is defined here instead of in root so that we can define an appropriate default.
	brokersCmd.PersistentFlags().StringVar(&brokersFlags.host, flags.HostLongName, flags.HostBrokerDefault, "")
	brokersCmd.PersistentFlags().MarkHidden(flags.HostLongName)

	// Flags for brokers create.
	flags.StringFlag(brokersCreateCmd.PersistentFlags(), &brokersFlags.broker, flags.BrokerLongName, flags.BrokerShortName, "[Required] Name of broker to create.")
	flags.StringFlag(brokersCreateCmd.PersistentFlags(), &brokersFlags.title, "title", "t", "[Optional] Title of broker to create. Defaults to broker name")

	// Flags for brokers delete.
	flags.StringFlag(brokersDeleteCmd.PersistentFlags(), &brokersFlags.broker, flags.BrokerLongName, flags.BrokerShortName, "[Required] The name of the broker.")

	// Flags for brokers cleanup.
	flags.StringFlag(brokersCleanupCmd.PersistentFlags(), &brokersFlags.broker, flags.BrokerLongName, flags.BrokerShortName, "[Required] The name of the broker.")
	flags.BoolFlag(brokersCleanupCmd.PersistentFlags(), &brokersFlags.verbose, "verbose", "v",
		"[Optional] If specified, the tool will print verbose logs indicating progress. (Default: FALSE)")

	RootCmd.AddCommand(brokersCmd)
	// TODO: Uncomment brokerListCmd when ListBroker is implmeneted in SB API.
	brokersCmd.AddCommand(brokersCreateCmd, brokersDeleteCmd, brokersCleanupCmd, brokersListCmd)
}

func cleanupBroker(brokerURL string) error {
	client := httpAdapterFromFlag()
	showProgress := brokersFlags.verbose
	lir, err := client.ListInstances(&adapter.ListInstancesParams{
		Server: brokerURL,
	})
	if err != nil {
		return err
	}

	errorMap := make(map[string]error)
	for _, i := range lir.Instances {
		lbr, err := client.ListBindings(&adapter.ListBindingsParams{
			Server:     brokerURL,
			InstanceID: i.ID,
		})
		if err != nil {
			errorMap[i.ID] = err
			continue
		}

		for _, b := range lbr.Bindings {
			if err := deleteBinding(client, flags.ApiVersionDefault, brokerURL, i, b, showProgress); err != nil {
				errorMap[i.ID] = err
				break
			}
		}

		if _, ok := errorMap[i.ID]; !ok {
			if err := deleteInstance(client, flags.ApiVersionDefault, brokerURL, i, showProgress); err != nil {
				errorMap[i.ID] = err
			}
		}
	}

	if len(errorMap) > 0 {
		return fmt.Errorf("Failed to cleanup service instances in broker with error %v", errorMap)
	}

	return nil
}

func printListBrokers(result *adapter.ListBrokersResult) {
	for index, b := range result.Brokers {
		fmt.Printf("%d. %s\n", index+1, b.Name)
		fmt.Printf("   URL: %s\n", *b.URL)
		fmt.Printf("   Create time: %s\n\n", *b.CreateTime)
	}
}
