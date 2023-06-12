package model

type ResourceAccess map[string]map[string]Access

type Access uint8

type RakeesMetrics struct {
	Name        string
	Create      string
	Delete      string
	List        string
	Update      string
	ClusterName string
}
