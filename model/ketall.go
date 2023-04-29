package model

type Resource struct {
	Resource    string `json:"resource"`
	Namespace   string `json:"namespace"`
	Age         string `json:"age"`
	ClusterName string `json:"clustername"`
}
