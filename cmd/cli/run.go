package cli

import (
	"errors"
	"github.com/rhsyseng/cluster0-operators/pkg/run"
	"github.com/spf13/cobra"
	"os"
)

var (
	kubeconfigFile                string
	operatorsInstallManifestsPath string
	operatorsConfigManifestsPath  string
)

func NewRunCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "run",
		Short:        "Waits for the cluster to be fully installed and deploys the operators and operator configs",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Validate command Args
			err := validateRunCommandArgs()
			if err != nil {
				return err
			}
			// We have the run command logic implemented in our example pkg
			err = run.RunCommandRun(kubeconfigFile, operatorsInstallManifestsPath, operatorsConfigManifestsPath)
			if err != nil {
				return err
			}
			return err
		},
	}
	addRunCommandFlags(cmd)
	return cmd
}

func addRunCommandFlags(cmd *cobra.Command) {

	flags := cmd.Flags()
	flags.StringVarP(&kubeconfigFile, "kubeconfig", "k", "", "Path to the kubeconfig file to be used. If not set, will default to in-cluster auth")
	flags.StringVarP(&operatorsInstallManifestsPath, "operators-install-manifests", "i", "", "Path to the folder where manifests for the operators to be installed are present")
	flags.StringVarP(&operatorsConfigManifestsPath, "operators-config-manifests", "c", "", "Path to the folder where manifests for the operators configurations to be applied are present")
	cmd.MarkFlagRequired("operators-install-manifests")
	cmd.MarkFlagRequired("operators-config-manifests")
}

// validateCommandArgs validates that arguments passed by the user are valid
func validateRunCommandArgs() error {
	if _, err := os.Stat(operatorsInstallManifestsPath); err != nil {
		return errors.New("Operators install manifests path " + operatorsInstallManifestsPath + " does not exist.")
	}
	if _, err := os.Stat(operatorsConfigManifestsPath); err != nil {
		return errors.New("Operators config manifests path " + operatorsConfigManifestsPath + " does not exist.")
	}
	if kubeconfigFile != "" {
		if _, err := os.Stat(kubeconfigFile); err != nil {
			return errors.New("Kubeconfig file " + kubeconfigFile + " does not exist.")
		}
	}

	return nil
}
