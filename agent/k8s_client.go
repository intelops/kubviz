package main

import (
	"gopkg.in/yaml.v2"
	metav1 "sigs.k8s.io/kustomize/pseudo/k8s/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kustomize/pseudo/k8s/client-go/kubernetes"
	"sigs.k8s.io/kustomize/pseudo/k8s/client-go/rest"
	"sigs.k8s.io/kustomize/pseudo/k8s/client-go/tools/clientcmd"
	clientcmdapi "sigs.k8s.io/kustomize/pseudo/k8s/client-go/tools/clientcmd/api"
)

type K8sData struct {
	Namespace          string
	ServiceAccountName string
	KubeconfigFileName string
}

func inclusterConfig() (*rest.Config, error) {
	restConfig, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	return restConfig, err
}

func outclusterConfig() (*rest.Config, error) {
	//var kubeconfig *string
	//if home := homeDir(); home != "" {
	//	kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	//} else {
	//	kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	//}
	//flag.Parse()

	// use the current context in kubeconfig
	restConfig, err := clientcmd.BuildConfigFromFlags("", "/Users/Avinash/.kube/config")
	if err != nil {
		panic(err.Error())
	}
	return restConfig, err
}

func getRestConfig() (*rest.Config, error) {
	// creates the in-cluster restConfig
	restConfig, err := inclusterConfig()

	// creates the out-cluster restConfig
	//restConfig, err := outclusterConfig()
	return restConfig, err
}

func (k8sData *K8sData) GetClientSet() (*kubernetes.Clientset, *rest.Config, error) {
	restConfig, err := getRestConfig()
	if err != nil {
		panic(err.Error())
	}
	// create the clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		panic(err.Error())
	}
	return clientset, restConfig, err
}

func (k8sData *K8sData) GenerateKubeConfiguration() (string, error) {
	// Get secret list for namespace
	clientset, restConfig, err := k8sData.GetClientSet()
	if err != nil {
		return "", err
	}
	sa, err := clientset.CoreV1().ServiceAccounts(k8sData.Namespace).Get(k8sData.ServiceAccountName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	secretName := sa.Secrets[0].Name
	secret, err := clientset.CoreV1().Secrets(k8sData.Namespace).Get(secretName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	clusters := make(map[string]*clientcmdapi.Cluster)
	clusters["default-cluster"] = &clientcmdapi.Cluster{
		Server:                   restConfig.Host,
		CertificateAuthorityData: secret.Data["ca.crt"],
	}

	contexts := make(map[string]*clientcmdapi.Context)
	contexts["default-context"] = &clientcmdapi.Context{
		Cluster:   "default-cluster",
		Namespace: k8sData.Namespace,
		AuthInfo:  k8sData.Namespace,
	}

	authinfos := make(map[string]*clientcmdapi.AuthInfo)
	authinfos[k8sData.Namespace] = &clientcmdapi.AuthInfo{
		Token: string(secret.Data["token"]),
	}

	clientConfig := clientcmdapi.Config{
		Kind:           "Config",
		APIVersion:     "v1",
		Clusters:       clusters,
		Contexts:       contexts,
		CurrentContext: "default-context",
		AuthInfos:      authinfos,
	}
	clientcmd.WriteToFile(clientConfig, k8sData.KubeconfigFileName)
	output, err := yaml.Marshal(clientConfig)
	return string(output), nil
}
