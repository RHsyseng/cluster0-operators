package run

import (
	"errors"
	"github.com/TwiN/go-color"
	"github.com/mvazquezc/cluster0-operators/pkg/olm"
	"github.com/mvazquezc/cluster0-operators/pkg/utils"
	"log"
)

func RunCommandRun(waitingTime int, kubeconfigFile string, operatorsInstallManifestsPath string, operatorsConfigManifestsPath string) error {
	log.Printf(color.InBlue("[INFO] ") + "Cluster0 operators deployer started\n")
	// Verify manifests folder have yaml files
	operatorInstallManifests := utils.ReturnManifestInFolder(operatorsInstallManifestsPath)
	operatorConfigManifests := utils.ReturnManifestInFolder(operatorsConfigManifestsPath)
	if len(operatorInstallManifests) <= 0 {
		return errors.New("No YAML files found at " + operatorsInstallManifestsPath)
	}
	if len(operatorConfigManifests) <= 0 {
		return errors.New("No YAML files found at " + operatorsConfigManifestsPath)
	}
	// Get a kubeclient
	client, dClient, err := utils.NewKubeClients(kubeconfigFile)
	if err != nil {
		return err
	}
	// Check CVO status
	err = olm.WaitForCVO(client)
	if err != nil {
		return err
	}
	// Install operators
	log.Printf(color.InBlue("[INFO] ") + "About to install operators\n")
	err = olm.ApplyManifestsInFolder(client, dClient, operatorInstallManifests)
	if err != nil {
		return err
	}
	log.Printf(color.InBlue("[INFO] ") + "About to configure operators\n")
	err = olm.ApplyManifestsInFolder(client, dClient, operatorConfigManifests)
	if err != nil {
		return err
	}
	log.Printf(color.InBlue("[INFO] ") + "Cluster0 operators deployer finished\n")
	return nil
}
