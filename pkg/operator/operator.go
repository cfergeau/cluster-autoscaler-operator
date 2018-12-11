package operator

import (
	"github.com/openshift/cluster-autoscaler-operator/pkg/apis"
	"github.com/openshift/cluster-autoscaler-operator/pkg/controller/clusterautoscaler"
	"github.com/openshift/cluster-autoscaler-operator/pkg/controller/machineautoscaler"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

// Operator represents an instance of the cluster-autoscaler-operator.
type Operator struct {
	config  *Config
	manager manager.Manager
}

// New returns a new Operator instance with the given config and a
// manager configured with the various controllers.
func New(cfg *Config) (*Operator, error) {
	operator := &Operator{config: cfg}

	// Get a config to talk to the apiserver.
	clientConfig, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	managerOptions := manager.Options{Namespace: cfg.WatchNamespace}
	operator.manager, err = manager.New(clientConfig, managerOptions)
	if err != nil {
		return nil, err
	}

	// Setup Scheme for all resources.
	if err := apis.AddToScheme(operator.manager.GetScheme()); err != nil {
		return nil, err
	}

	if err := operator.AddControllers(); err != nil {
		return nil, err
	}

	return operator, nil
}

// AddControllers configures the various controllers and adds them to
// the operator's manager instance.
func (o *Operator) AddControllers() error {
	// Setup ClusterAutoscaler controller.
	ca := clusterautoscaler.NewReconciler(o.manager, &clusterautoscaler.Config{
		Name:      o.config.ClusterAutoscalerName,
		Image:     o.config.ClusterAutoscalerImage,
		Replicas:  o.config.ClusterAutoscalerReplicas,
		Namespace: o.config.ClusterAutoscalerNamespace,
	})

	if err := ca.AddToManager(o.manager); err != nil {
		return err
	}

	// Setup MachineAutoscaler controller.
	ma := machineautoscaler.NewReconciler(o.manager)
	if err := ma.AddToManager(o.manager); err != nil {
		return err
	}

	return nil
}

// Start starts the operator's controller-manager.
func (o *Operator) Start() error {
	return o.manager.Start(signals.SetupSignalHandler())
}
