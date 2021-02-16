package k8sheartbeat

import (
	"context"
	"errors"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

var (
	ErrPodUnavailable        = errors.New("the given pod is unavailable")
	ErrMinPodsNotAvailable   = errors.New("there should be at least one pod available")
	ErrDeploymentUnavailable = errors.New("deployment is unavailable")
	ErrPodEventsUnhealthy    = errors.New("pod has unhealthy events")
)

// HealthChecker defines a health check interface
type K8sHealthChecker struct {
	// Cluster namespace to use
	Namespace string
	// Kubernetes client
	ClientSet *kubernetes.Clientset
}

func NewK8sHealthChecker(namespace string, clientset *kubernetes.Clientset) *K8sHealthChecker {
	return &K8sHealthChecker{Namespace: namespace, ClientSet: clientset}
}

// HealthCheck is a HealthChecker implementation for a kubernetes deployment
func (k *K8sHealthChecker) HealthCheck(ctx context.Context, name string) error {
	err := k.isComponentHealthy(ctx, name)
	if err != nil {
		return err
	}

	return nil
}

// Ping checks if the K8S api is available
func (k *K8sHealthChecker) Ping() error {
	_, err := k.ClientSet.RESTClient().Get().AbsPath("/healthz").DoRaw(context.Background())
	return err
}

// In order to ensure a deployment is working properly we should state the following:
// - Check deployment's condition for Type Available, should have status True
// - Randomly choose a single pod in the deployment, check if the last event is:
// 	- Type: Warning, Reason: Backoff
// 	- Type: Warning, Reason: Unhealthy
func (k *K8sHealthChecker) isComponentHealthy(ctx context.Context, name string) error {
	deployment, err := k.ClientSet.AppsV1().Deployments(k.Namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	pods, err := k.getDeploymentPods(ctx, deployment)
	if err != nil {
		return err
	}

	log.Info("Pods are: ", pods.Items)

	lenPods := len(pods.Items)
	if lenPods <= 0 {
		return ErrMinPodsNotAvailable
	}

	rand.Seed(time.Now().UnixNano())
	pod := pods.Items[rand.Intn(lenPods)]
	log.Info("Pod: ", pod.Name)

	if !isPodAvailable(&pod) {
		return ErrPodUnavailable
	}

	podHasHealthyEvents, err := k.isPodEventsHealthy(ctx, &pod)
	if err != nil {
		return err
	}

	if !podHasHealthyEvents {
		return ErrPodEventsUnhealthy
	}

	return nil
}

func (k *K8sHealthChecker) getDeploymentPods(ctx context.Context, deployment *appsv1.Deployment) (*corev1.PodList, error) {
	labelsMap, err := metav1.LabelSelectorAsMap(deployment.Spec.Selector)
	if err != nil {
		return nil, err
	}
	set := labels.Set(labelsMap)

	listOptions := metav1.ListOptions{LabelSelector: set.AsSelector().String()}
	return k.ClientSet.CoreV1().Pods(k.Namespace).List(ctx, listOptions)
}

func (k *K8sHealthChecker) getPodEvents(ctx context.Context, pod *corev1.Pod) (*corev1.EventList, error) {
	listOptions := metav1.ListOptions{FieldSelector: fields.OneTermEqualSelector("involvedObject.name", pod.Name).String()}
	return k.ClientSet.CoreV1().Events(k.Namespace).List(ctx, listOptions)
}

func (k *K8sHealthChecker) isPodEventsHealthy(ctx context.Context, pod *corev1.Pod) (bool, error) {
	events, err := k.getPodEvents(ctx, pod)
	if err != nil {
		return false, err
	}

	lenEvents := len(events.Items)
	if lenEvents <= 0 {
		// Normally long living pods have no events
		log.Info("Pod has no events")
		return true, nil
	}

	event := events.Items[len(events.Items)-1]
	log.Debug("Event Type: ", event.Type)
	log.Debug("Event Reason: ", event.Reason)
	log.Debug("Event Message: ", event.Message)

	switch event.Type {
	case "Warning":
	case "Error":
		return false, nil
	}

	return true, nil
}

func isPodAvailable(pod *corev1.Pod) bool {
	for _, c := range pod.Status.Conditions {
		if c.Type == corev1.PodReady {
			return c.Status == corev1.ConditionTrue
		}
	}

	return false
}

func isDeploymentAvailable(deployment *appsv1.Deployment) bool {
	for _, c := range deployment.Status.Conditions {
		if c.Type == appsv1.DeploymentAvailable {
			return c.Status == corev1.ConditionTrue
		}
	}

	return false
}
