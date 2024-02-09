package model

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type State struct {
	OK            bool                       `json:"OK"`
	Errors        []string                   `json:"Errors"`
	CheckDetails  map[string]WorkloadDetails `json:"CheckDetails"` // map of check names to last run timestamp
	JobDetails    map[string]WorkloadDetails `json:"JobDetails"`   // map of job names to last run timestamp
	CurrentMaster string                     `json:"CurrentMaster"`
	Metadata      map[string]string          `json:"Metadata"`
}

type WorkloadDetails struct {
	OK               bool         `json:"OK" yaml:"OK"`                               // true or false status of the khWorkload, whether or not it completed successfully
	Errors           []string     `json:"Errors" yaml:"Errors"`                       // the list of errors reported from the khWorkload run
	RunDuration      string       `json:"RunDuration" yaml:"RunDuration"`             // the time it took for the khWorkload to complete
	Namespace        string       `json:"Namespace" yaml:"Namespace"`                 // the namespace the khWorkload was run in
	Node             string       `json:"Node" yaml:"Node"`                           // the node the khWorkload ran on
	LastRun          *metav1.Time `json:"LastRun,omitempty" yaml:"LastRun,omitempty"` // the time the khWorkload was last run
	AuthoritativePod string       `json:"AuthoritativePod" yaml:"AuthoritativePod"`   // the main kuberhealthy pod creating and updating the khstate
	CurrentUUID      string       `json:"uuid" yaml:"uuid"`                           // the UUID that is authorized to report statuses into the kuberhealthy endpoint
}

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
