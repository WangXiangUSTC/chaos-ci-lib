package tests

import (
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/litmuschaos/chaos-operator/pkg/apis/litmuschaos/v1alpha1"
	chaosClient "github.com/litmuschaos/chaos-operator/pkg/client/clientset/versioned/typed/litmuschaos/v1alpha1"
	"github.com/mayadata-io/chaos-ci-lib/pkg"
	chaosTypes "github.com/mayadata-io/chaos-ci-lib/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	scheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog"
)

func TestPodNetworkLatency(t *testing.T) {

	RegisterFailHandler(Fail)
	RunSpecs(t, "BDD test")
}

//BDD Tests for pod-network-latency experiment
var _ = Describe("BDD of pod-network-latency experiment", func() {

	// BDD TEST CASE 1
	Context("Check for Litmus components", func() {

		It("Should check for creation of runner pod", func() {

			var err error
			var experimentName = "pod-network-latency"
			var engineName = "engine7"
			//Prerequisite of the test
			chaosTypes.Config, err = pkg.GetKubeConfig()
			if err != nil {
				Expect(err).To(BeNil(), "Failed to get kubeconfig client")
			}
			chaosTypes.Client, err = kubernetes.NewForConfig(chaosTypes.Config)
			if err != nil {
				Expect(err).To(BeNil(), "failed to get client")
			}
			chaosTypes.ClientSet, err = chaosClient.NewForConfig(chaosTypes.Config)
			if err != nil {
				Expect(err).To(BeNil(), "failed to get clientSet")
			}
			err = v1alpha1.AddToScheme(scheme.Scheme)
			if err != nil {
				fmt.Println(err)
			}
			//Installing RBAC for the experiment
			err = pkg.InstallRbac(chaosTypes.PodNetworkLatencyRbacPath, pkg.GetEnv("APP_NS", "default"), experimentName, chaosTypes.Client)
			Expect(err).To(BeNil(), "Fail to create RBAC")
			klog.Info("Rbac has been created successfully !!!")

			//Installing chaos engine for the experiment
			//Fetching engine file
			By("Fetching engine file for the experiment")
			err = pkg.DownloadFile(experimentName+"-ce.yaml", chaosTypes.PodNetworkLatencyEnginePath)
			Expect(err).To(BeNil(), "Fail to fetch engine file")

			//Modify chaos engine spec
			err = pkg.EditFile(experimentName+"-ce.yaml", "name: nginx-network-chaos", "name: "+engineName)
			Expect(err).To(BeNil(), "Failed to update engine name in engine")
			err = pkg.EditFile(experimentName+"-ce.yaml", "namespace: default", "namespace: "+pkg.GetEnv("APP_NS", "default"))
			Expect(err).To(BeNil(), "Failed to update namespace in engine")
			err = pkg.EditFile(experimentName+"-ce.yaml", "appns: 'default'", "appns: '"+pkg.GetEnv("APP_NS", "default")+"'")
			Expect(err).To(BeNil(), "Failed to update application namespace in engine")
			err = pkg.EditFile(experimentName+"-ce.yaml", "annotationCheck: 'true'", "annotationCheck: 'false'")
			Expect(err).To(BeNil(), "Failed to update AnnotationCheck in engine")
			err = pkg.EditFile(experimentName+"-ce.yaml", "applabel: 'app=nginx'", "applabel: '"+pkg.GetEnv("APP_LABEL", "run=nginx")+"'")
			Expect(err).To(BeNil(), "Failed to update application label in engine")
			err = pkg.EditFile(experimentName+"-ce.yaml", "appkind: 'deployment'", "appkind: "+pkg.GetEnv("APP_KIND", "deployment"))
			Expect(err).To(BeNil(), "Failed to update application kind in engine")
			err = pkg.EditFile(experimentName+"-ce.yaml", "jobCleanUpPolicy: 'delete'", "jobCleanUpPolicy: 'retain'")
			Expect(err).To(BeNil(), "Failed to update application label in engine")
			err = pkg.EditFile(experimentName+"-ce.yaml", "name: pod-network-latency", "name: k8s-pod-network-latency")
			Expect(err).To(BeNil(), "Failed to update the chaosexperiment name")
			err = pkg.EditKeyValue(experimentName+"-ce.yaml", "TOTAL_CHAOS_DURATION", "value: '30'", "value: '"+pkg.GetEnv("TOTAL_CHAOS_DURATION", "60")+"'")
			Expect(err).To(BeNil(), "Failed to update total chaos duration")
			err = pkg.EditKeyValue(experimentName+"-ce.yaml", "NETWORK_INTERFACE", "value: ''", "value: '"+pkg.GetEnv("NETWORK_INTERFACE", "eth0")+"'")
			Expect(err).To(BeNil(), "Failed to update the network interface")
			err = pkg.EditKeyValue(experimentName+"-ce.yaml", "TARGET_CONTAINER", "value: 'nginx'", "value: '"+pkg.GetEnv("TARGET_CONTAINER", "nginx")+"'")
			Expect(err).To(BeNil(), "Failed to update the target container name")
			err = pkg.EditKeyValue(experimentName+"-ce.yaml", "NETWORK_LATENCY", "value: 'nginx'", "value: '"+pkg.GetEnv("NETWORK_LATENCY", "60000")+"'")
			Expect(err).To(BeNil(), "Failed to update the network latency value")

			//Creating ChaosEngine
			By("Creating ChaosEngine")
			err = exec.Command("kubectl", "apply", "-f", experimentName+"-ce.yaml", "-n", pkg.GetEnv("APP_NS", "default")).Run()
			Expect(err).To(BeNil(), "Fail to create ChaosEngine")
			klog.Info("ChaosEngine created successfully")
			time.Sleep(2 * time.Second)

			//Fetching the runner pod and Checking if it get in Running state or not
			By("Wait for runner pod to come in running sate")
			runnerNamespace := pkg.GetEnv("APP_NS", "default")
			runnerPodStatus, err := pkg.RunnerPodStatus(runnerNamespace, engineName, chaosTypes.Client)
			Expect(runnerPodStatus).NotTo(Equal("1"), "Runner pod failed to get in running state")
			Expect(err).To(BeNil(), "Fail to get the runner pod")
			klog.Info("Runner pod for is in Running state")

			//Waiting for experiment job to get completed
			//Also Printing the logs of the experiment
			By("Waiting for job completion")
			jobNamespace := pkg.GetEnv("APP_NS", "default")
			jobPodLogs, err := pkg.JobLogs(experimentName, jobNamespace, engineName, chaosTypes.Client)
			Expect(jobPodLogs).To(Equal(0), "Fail to print the logs of the experiment")
			Expect(err).To(BeNil(), "Fail to get the experiment job pod")

			//Checking the chaosresult
			By("Checking the chaosresult")
			app, err := chaosTypes.ClientSet.ChaosResults(pkg.GetEnv("APP_NS", "default")).Get(engineName+"-"+experimentName, metav1.GetOptions{})
			Expect(string(app.Status.ExperimentStatus.Verdict)).To(Equal("Pass"), "Verdict is not pass chaosresult")
			Expect(err).To(BeNil(), "Fail to get chaosresult")
		})
	})
})
