package graph

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/intelops/kubviz/graphqlserver/graph/model"
)

func (r *Resolver) fetchNamespacesFromDatabase(ctx context.Context) ([]string, error) {
	if r.DB == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}
	query := `SELECT DISTINCT Namespace FROM events`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	var namespaces []string
	for rows.Next() {
		var namespace string
		if err := rows.Scan(&namespace); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		namespaces = append(namespaces, namespace)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return namespaces, nil
}
func (r *Resolver) fetchOutdatedImages(ctx context.Context, namespace string) ([]*model.OutdatedImage, error) {
	if r.DB == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}
	query := `SELECT ClusterName, Namespace, Pod, CurrentImage, CurrentTag, LatestVersion, VersionsBehind, EventTime FROM outdated_images WHERE Namespace = ?`

	rows, err := r.DB.QueryContext(ctx, query, namespace)
	if err != nil {
		if err == sql.ErrNoRows {
			return []*model.OutdatedImage{}, nil
		}
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	var outdatedImages []*model.OutdatedImage
	for rows.Next() {
		var oi model.OutdatedImage
		if err := rows.Scan(&oi.ClusterName, &oi.Namespace, &oi.Pod, &oi.CurrentImage, &oi.CurrentTag, &oi.LatestVersion, &oi.VersionsBehind, &oi.EventTime); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		outdatedImages = append(outdatedImages, &oi)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return outdatedImages, nil
}
func (r *Resolver) fetchKubeScores(ctx context.Context, namespace string) ([]*model.KubeScore, error) {
	if r.DB == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	query := `SELECT id, clustername, object_name, kind, apiVersion, name, namespace, target_type, description, path, summary, file_name, file_row, EventTime FROM kubescore WHERE namespace = ?`
	rows, err := r.DB.QueryContext(ctx, query, namespace)
	if err != nil {
		if err == sql.ErrNoRows {
			// No data for the namespace, return an empty slice
			return []*model.KubeScore{}, nil
		}
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	var kubeScores []*model.KubeScore
	for rows.Next() {
		var ks model.KubeScore
		if err := rows.Scan(&ks.ID, &ks.ClusterName, &ks.ObjectName, &ks.Kind, &ks.APIVersion, &ks.Name, &ks.Namespace, &ks.TargetType, &ks.Description, &ks.Path, &ks.Summary, &ks.FileName, &ks.FileRow, &ks.EventTime); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		kubeScores = append(kubeScores, &ks)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return kubeScores, nil
}
func (r *Resolver) fetchResources(ctx context.Context, namespace string) ([]*model.Resource, error) {
	if r.DB == nil {
		return nil, fmt.Errorf("database connection is not initialized")
	}

	query := `SELECT ClusterName, Namespace, Kind, Resource, Age, EventTime FROM getall_resources WHERE Namespace = ?`
	rows, err := r.DB.QueryContext(ctx, query, namespace)
	if err != nil {
		if err == sql.ErrNoRows {
			// No data for the namespace, return an empty slice
			return []*model.Resource{}, nil
		}
		return nil, fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	var resources []*model.Resource
	for rows.Next() {
		var res model.Resource
		if err := rows.Scan(&res.ClusterName, &res.Namespace, &res.Kind, &res.Resource, &res.Age, &res.EventTime); err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}
		resources = append(resources, &res)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return resources, nil
}
