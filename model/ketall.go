package model

type Resource struct {
	Resource    string `json:"resource"`
	Kind        string `json:"kind"`
	Namespace   string `json:"namespace"`
	Age         string `json:"age"`
	ClusterName string `json:"clustername"`
}
