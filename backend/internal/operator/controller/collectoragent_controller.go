package controller

import (
	"context"
	"time"

	"github.com/neuralops/platform/internal/operator/api/v1"
	"github.com/neuralops/platform/internal/operator/gateway"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// CollectorAgentReconciler syncs CollectorAgent CRs to the gateway fleet API.
type CollectorAgentReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	Gateway       *gateway.Client
	RequeueAfter  time.Duration
}

// +kubebuilder:rbac:groups=observability.neuralops.io,resources=collectoragents,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=observability.neuralops.io,resources=collectoragents/status,verbs=get;update;patch

func (r *CollectorAgentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	var agent v1.CollectorAgent
	if err := r.Get(ctx, req.NamespacedName, &agent); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	gw := r.Gateway
	if gw == nil {
		gw = gateway.NewClientFromEnv()
	}

	agentID := agent.Name
	if agentID == "" {
		agentID = agent.Spec.Name
	}
	env := agent.Spec.Environment
	if env == "" {
		env = "prod"
	}
	policy := agent.Spec.PolicyID
	if policy == "" {
		policy = "default"
	}

	phase := v1.PhaseSynced
	message := "synced to gateway"
	targetVersion := agent.Spec.Version

	if agent.Spec.CanaryVersion != "" && agent.Status.ObservedVersion != agent.Spec.CanaryVersion {
		if err := gw.UpgradeAgent(ctx, agentID, agent.Spec.CanaryVersion); err != nil {
			phase = v1.PhaseFailed
			message = err.Error()
		} else {
			phase = v1.PhaseUpgrading
			targetVersion = agent.Spec.CanaryVersion
			message = "canary upgrade applied"
		}
	} else if err := gw.SyncAgent(ctx, agentID, agent.Spec.Name, env, agent.Spec.Version, policy, ""); err != nil {
		phase = v1.PhaseFailed
		message = err.Error()
		logger.Error(err, "gateway sync failed", "agent", agentID)
	}

	agent.Status.Phase = phase
	agent.Status.Message = message
	agent.Status.ObservedVersion = targetVersion
	agent.Status.GatewayAgentID = agentID
	agent.Status.LastReconcileTime = metav1.NewTime(time.Now().UTC())
	if err := r.Status().Update(ctx, &agent); err != nil {
		return ctrl.Result{}, err
	}

	requeue := r.RequeueAfter
	if requeue <= 0 {
		requeue = 30 * time.Second
	}
	return ctrl.Result{RequeueAfter: requeue}, nil
}

// SetupWithManager registers the reconciler.
func (r *CollectorAgentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.CollectorAgent{}).
		Complete(r)
}
