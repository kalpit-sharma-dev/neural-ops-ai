package collector

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// K8sAPIClient syncs inventory from the Kubernetes API (in-cluster or KUBECONFIG host).
type K8sAPIClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
	pool       *pgxpool.Pool
	tenant     string
}

// NewK8sAPIClient builds a client from in-cluster service account or env.
func NewK8sAPIClient(pool *pgxpool.Pool, tenantID string) *K8sAPIClient {
	if tenantID == "" {
		tenantID = "default"
	}
	host := os.Getenv("KUBERNETES_SERVICE_HOST")
	port := os.Getenv("KUBERNETES_SERVICE_PORT")
	baseURL := os.Getenv("KUBERNETES_API_URL")
	token := ""
	if host != "" && port != "" && baseURL == "" {
		baseURL = fmt.Sprintf("https://%s:%s", host, port)
		tokenBytes, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
		if err == nil {
			token = strings.TrimSpace(string(tokenBytes))
		}
	}
	if baseURL == "" {
		return nil
	}
	return &K8sAPIClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // in-cluster default CA
			},
		},
		pool:   pool,
		tenant: tenantID,
	}
}

// Available reports whether K8s API sync can run.
func (k *K8sAPIClient) Available() bool {
	return k != nil && k.baseURL != "" && k.pool != nil
}

type k8sListMeta struct {
	Items []json.RawMessage `json:"items"`
}

type k8sObjectMeta struct {
	Metadata struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
	Status struct {
		Phase string `json:"phase"`
	} `json:"status"`
	Spec struct {
		NodeName string `json:"nodeName"`
	} `json:"spec"`
}

type k8sDeployment struct {
	Metadata struct {
		Name      string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
	Spec struct {
		Replicas int32 `json:"replicas"`
	} `json:"spec"`
	Status struct {
		ReadyReplicas int32 `json:"readyReplicas"`
	} `json:"status"`
}

// SyncNamespacesPodsDeployments pulls live cluster inventory.
func (k *K8sAPIClient) Sync(ctx context.Context) error {
	if !k.Available() {
		return fmt.Errorf("k8s api unavailable")
	}
	nsCount, err := k.syncNamespaces(ctx)
	if err != nil {
		return err
	}
	podCount, err := k.syncPods(ctx)
	if err != nil {
		return err
	}
	if err := k.syncDeployments(ctx); err != nil {
		return err
	}
	_, _ = k.pool.Exec(ctx, `
INSERT INTO collector_k8s_clusters (tenant_id, name, nodes, pods, health, namespace_count, updated_at)
VALUES ($1, 'production', 0, $2, 'healthy', $3, NOW())
ON CONFLICT (tenant_id, name) DO UPDATE SET
  pods = EXCLUDED.pods, namespace_count = EXCLUDED.namespace_count, updated_at = NOW(), health = 'healthy'`,
		k.tenant, podCount, nsCount,
	)
	return nil
}

func (k *K8sAPIClient) syncNamespaces(ctx context.Context) (int, error) {
	body, err := k.get(ctx, "/api/v1/namespaces")
	if err != nil {
		return 0, err
	}
	var list k8sListMeta
	if err := json.Unmarshal(body, &list); err != nil {
		return 0, err
	}
	count := 0
	for _, raw := range list.Items {
		var obj k8sObjectMeta
		if json.Unmarshal(raw, &obj) != nil || obj.Metadata.Name == "" {
			continue
		}
		count++
		_, _ = k.pool.Exec(ctx, `
INSERT INTO collector_k8s_namespaces (tenant_id, name, status, pod_count, updated_at)
VALUES ($1, $2, 'Active', 0, NOW())
ON CONFLICT (tenant_id, name) DO UPDATE SET updated_at = NOW()`,
			k.tenant, obj.Metadata.Name)
	}
	return count, nil
}

func (k *K8sAPIClient) syncPods(ctx context.Context) (int, error) {
	body, err := k.get(ctx, "/api/v1/pods")
	if err != nil {
		return 0, err
	}
	var list k8sListMeta
	if err := json.Unmarshal(body, &list); err != nil {
		return 0, err
	}
	count := 0
	for _, raw := range list.Items {
		var obj k8sObjectMeta
		if json.Unmarshal(raw, &obj) != nil || obj.Metadata.Name == "" {
			continue
		}
		ns := obj.Metadata.Namespace
		if ns == "" {
			ns = "default"
		}
		phase := obj.Status.Phase
		if phase == "" {
			phase = "Running"
		}
		node := obj.Spec.NodeName
		count++
		_, _ = k.pool.Exec(ctx, `
INSERT INTO collector_k8s_pods (tenant_id, name, namespace, node, status, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (tenant_id, namespace, name) DO UPDATE SET
  node = EXCLUDED.node, status = EXCLUDED.status, updated_at = NOW()`,
			k.tenant, obj.Metadata.Name, ns, node, phase)
	}
	return count, nil
}

func (k *K8sAPIClient) syncDeployments(ctx context.Context) error {
	body, err := k.get(ctx, "/apis/apps/v1/deployments")
	if err != nil {
		return err
	}
	var list k8sListMeta
	if err := json.Unmarshal(body, &list); err != nil {
		return err
	}
	for _, raw := range list.Items {
		var dep k8sDeployment
		if json.Unmarshal(raw, &dep) != nil || dep.Metadata.Name == "" {
			continue
		}
		ns := dep.Metadata.Namespace
		if ns == "" {
			ns = "default"
		}
		_, _ = k.pool.Exec(ctx, `
INSERT INTO collector_k8s_deployments (tenant_id, name, namespace, replicas, ready_replicas, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (tenant_id, namespace, name) DO UPDATE SET
  replicas = EXCLUDED.replicas, ready_replicas = EXCLUDED.ready_replicas, updated_at = NOW()`,
			k.tenant, dep.Metadata.Name, ns, dep.Spec.Replicas, dep.Status.ReadyReplicas)
	}
	return nil
}

func (k *K8sAPIClient) get(ctx context.Context, path string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, k.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	if k.token != "" {
		req.Header.Set("Authorization", "Bearer "+k.token)
	}
	resp, err := k.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("k8s api %s: %s", path, resp.Status)
	}
	return io.ReadAll(resp.Body)
}
