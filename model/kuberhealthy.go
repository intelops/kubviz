package model

import "time"

// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type KuberhealthyCheckDetail struct {
	CurrentUUID      string    `json:"currentUUID"`
	CheckName        string    `json:"checkName"`
	OK               uint8     `json:"ok"`
	Errors           string    `json:"errors"`
	RunDuration      string    `json:"runDuration"`
	Namespace        string    `json:"namespace"`
	Node             string    `json:"node"`
	LastRun          time.Time `json:"lastRun"`
	AuthoritativePod string    `json:"authoritativePod"`
}
