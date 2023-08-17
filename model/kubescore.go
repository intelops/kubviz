package model

import "github.com/zegl/kube-score/renderer/json_v2"

type KubeScoreRecommendations struct {
	ID          string
	ClusterName string
	Report      []json_v2.ScoredObject
}
