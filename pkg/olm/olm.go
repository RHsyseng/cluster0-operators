package olm

import (
	"context"
	"errors"
	color "github.com/TwiN/go-color"
	"github.com/mvazquezc/cluster0-operators/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"log"
	"strings"
	"time"
)

const (
	CVO_MAX_WAIT_SECONDS   = 7200
	CVO_WAIT_SLEEP_SECONDS = 30
	SUB_MAX_WAIT_SECONDS   = 300
	SUB_WAIT_SLEEP_SECONDS = 5
	STATUS_MAX_RETRIES     = 10
)

func ApplyManifestsInFolder(client *dynamic.DynamicClient, dClient *discovery.DiscoveryClient, manifestsFiles []string) error {

	var appliedObjects []unstructured.Unstructured
	//var err error
	for _, v := range manifestsFiles {
		log.Printf(color.InBlue("[INFO] ")+"Processing file %s\n", v)
		objs, err := utils.KubeClientApply(v, client, dClient)
		if err != nil {
			return err
		}
		//The ... syntax is used to spread the elements of objs into individual arguments to the append function
		appliedObjects = append(appliedObjects, objs...)
	}
	// We send the list of applied objects, this method will evaluate if we need to wait for something to happen
	// like wait for a Subscription to fully deploy the operator
	err := waitForManifests(client, appliedObjects)
	if err != nil {
		return err
	}
	return nil
}

func waitForManifests(client *dynamic.DynamicClient, appliedObjects []unstructured.Unstructured) error {
	for _, obj := range appliedObjects {
		if strings.ToLower(obj.GetKind()) == "subscription" {
			log.Printf(color.InBlue("[INFO] ")+"Detected Subscription %s in namespace %s, waiting for operator to be deployed\n", obj.GetName(), obj.GetNamespace())
			err := waitForSub(client, obj)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func waitForSub(client *dynamic.DynamicClient, obj unstructured.Unstructured) error {
	subRes := schema.GroupVersionResource{Group: "operators.coreos.com", Version: "v1alpha1", Resource: "subscriptions"}
	csvRes := schema.GroupVersionResource{Group: "operators.coreos.com", Version: "v1alpha1", Resource: "clusterserviceversions"}
	log.Printf(color.InBlue("[INFO] ")+"Waiting for Subscription %s. Timeout %ds, Wait Interval: %ds\n", obj.GetName(), SUB_MAX_WAIT_SECONDS, SUB_WAIT_SLEEP_SECONDS)
	waitTime := 0
	subStatusRetries := 0
	csvStatusRetries := 0
	for {
		subData, err := client.Resource(subRes).Namespace(obj.GetNamespace()).Get(context.TODO(), obj.GetName(), metav1.GetOptions{})
		if err != nil {
			return err
		}
		installedCSV, statusFound, err := unstructured.NestedString(subData.Object, "status", "installedCSV")
		if err != nil {
			return err
		}
		// Status may not be found, we need to count retries and exit the loop if we couldn't find the installedCSV after X iterations
		if statusFound {
			csvData, err := client.Resource(csvRes).Namespace(obj.GetNamespace()).Get(context.TODO(), installedCSV, metav1.GetOptions{})
			if err != nil {
				return err
			}
			csvPhase, phaseFound, err := unstructured.NestedString(csvData.Object, "status", "phase")
			if phaseFound {
				log.Printf(color.InBlue("[INFO] ")+"ClusterServiceVersion Name: %s, Phase: %s, Timeout: [%ds/%ds]\n", installedCSV, csvPhase, waitTime, SUB_MAX_WAIT_SECONDS)
				if strings.ToLower(csvPhase) == "succeeded" {
					log.Printf(color.InBlue("[INFO] ")+"Subscription %s completed\n", obj.GetName())
					break
				}
			} else {
				log.Printf(color.InBlue("[INFO] ")+"ClusterServiceVersion %s status doesn't have phase information yet. Retries: [%d/%d]\n", installedCSV, csvStatusRetries, STATUS_MAX_RETRIES)
				if csvStatusRetries >= STATUS_MAX_RETRIES {
					return errors.New("Timed out while waiting for ClusterServiceVersion to populate its status.phase field")
				}
				csvStatusRetries += 1
			}
		} else {
			log.Printf(color.InBlue("[INFO] ")+"Subscription %s status doesn't have installedCSV information yet. Retries: [%d/%d]\n", obj.GetName(), subStatusRetries, STATUS_MAX_RETRIES)
			if subStatusRetries >= STATUS_MAX_RETRIES {
				return errors.New("Timed out while waiting for Subscription to populate its status.installedCSV field")
			}
			subStatusRetries += 1
		}

		if waitTime >= SUB_MAX_WAIT_SECONDS {
			return errors.New("Timed out while waiting for Subscription to finish operator install")
		}
		time.Sleep(SUB_WAIT_SLEEP_SECONDS * time.Second)
		waitTime += SUB_WAIT_SLEEP_SECONDS
	}
	return nil
}

func WaitForCVO(client *dynamic.DynamicClient) error {
	cvoRes := schema.GroupVersionResource{Group: "config.openshift.io", Version: "v1", Resource: "clusterversions"}
	log.Printf(color.InBlue("[INFO] ")+"Waiting for CVO to finish. Timeout: %ds, Wait Interval: %ds\n", CVO_MAX_WAIT_SECONDS, CVO_WAIT_SLEEP_SECONDS)
	waitTime := 0
	for {
		cvoData, err := client.Resource(cvoRes).Get(context.TODO(), "version", metav1.GetOptions{})
		if err != nil {
			return err
		}
		cvoConditions, _, err := unstructured.NestedSlice(cvoData.Object, "status", "conditions")
		if err != nil {
			return err
		}
		for _, v := range cvoConditions {
			cvoConditionType := v.(map[string]interface{})["type"]
			cvoConditionStatus := v.(map[string]interface{})["status"]
			cvoConditionMessage := v.(map[string]interface{})["message"]
			if cvoConditionType == "Available" {
				log.Printf(color.InBlue("[INFO] ")+"CVO Finished: %v, Message: %s, Timeout: [%ds/%ds]\n", cvoConditionStatus, cvoConditionMessage, waitTime, CVO_MAX_WAIT_SECONDS)
				if cvoConditionStatus == "True" {
					return nil
				}
			}
		}
		if waitTime >= CVO_MAX_WAIT_SECONDS {
			return errors.New("Timed out while waiting for CVO to finish cluster install")
		}
		time.Sleep(CVO_WAIT_SLEEP_SECONDS * time.Second)
		waitTime += CVO_WAIT_SLEEP_SECONDS
	}
}
