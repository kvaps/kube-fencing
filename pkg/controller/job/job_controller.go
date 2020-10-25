package job

import (
	"context"
	"encoding/json"

	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new Job Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileJob{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("job-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Job
	err = c.Watch(&source.Kind{Type: &batchv1.Job{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileJob implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileJob{}

// ReconcileJob reconciles a Job object
type ReconcileJob struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Job object and makes changes based on the state read
// and what is in the Job.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileJob) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	// Fetch the Job instance
	instance := &batchv1.Job{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// We need only Jobs created by node controller
	if instance.Labels["fencing"] != "fence" {
		return reconcile.Result{}, err
	}

	// Ignore already fenced nodes
	if instance.Annotations["fencing/state"] == "fenced" {
		return reconcile.Result{}, nil
	}

	// Take the node name
	nodeName, ok := instance.Annotations["fencing/node"]
	if !ok {
		return reconcile.Result{}, err
	}

	// Get the node
	node := &v1.Node{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: nodeName}, node)
	if err != nil {
		klog.Errorln(err, "No node found", nodeName)
		return reconcile.Result{}, err
	}

	// We need to wait until job succeeded
	if instance.Status.Succeeded < 1 {
		return reconcile.Result{}, nil
		// Set fencing/status=error if job was failed
		if instance.Status.Failed > 0 {
			klog.Infoln("Failed fencing node", nodeName)
		}
	}

	klog.Infoln("Succesful fencing node", nodeName)

	// Get the fencing mode
	fencingMode, ok := instance.Annotations["fencing/mode"]
	if !ok {
		return reconcile.Result{}, nil
	}

	// Start the cleanup
	switch fencingMode {
	case "none":
		// Do nothing
	case "delete":
		// Delete the node
		klog.Infoln("Removing node", nodeName)
		err = r.client.Delete(context.TODO(), node)
		if err != nil {
			klog.Errorln("Failed to delete node", nodeName, ":", err)
			return reconcile.Result{}, nil
		}
	case "flush":
		// Flush all resources from the node
		klog.Infoln("Flushing node", nodeName)

		// Fetch a list of all namespaces for DeleteAllOf requests
		namespaces := v1.NamespaceList{}
		pod := &v1.Pod{}
		if err := r.client.List(context.TODO(), &namespaces); err != nil {
			klog.Errorln("Failed to get namespace list:", err)
		}
		for _, ns := range namespaces.Items {
			opts := []client.DeleteAllOfOption{
				client.InNamespace(ns.Name),
				client.MatchingFields{"spec.nodeName": nodeName},
				client.GracePeriodSeconds(0),
			}
			err = r.client.DeleteAllOf(context.TODO(), pod, opts...)
			if err != nil {
				klog.Errorln("Failed to delete pods in namespace", ns.Name, ":", err)
			}
		}

		// Fetch a list of all volumeattachments and delete them
		volumeattachment := &storagev1.VolumeAttachment{}
		volumeattachments := storagev1.VolumeAttachmentList{}
		if err := r.client.List(context.TODO(), &volumeattachments); err != nil {
			klog.Errorln("Failed to get volumeattachment list:", err)
		}
		for _, va := range volumeattachments.Items {
			if va.Spec.NodeName == nodeName {
				opts := []client.DeleteAllOfOption{
					client.MatchingFields{"metadata.name": va.Name},
					client.GracePeriodSeconds(0),
				}
				err = r.client.DeleteAllOf(context.TODO(), volumeattachment, opts...)
				if err != nil {
					klog.Errorln("Failed to delete volumeattachment", va.Name, ":", err)
				}
			}
		}
	default:
		klog.Errorln("Unknown fencing mode", fencingMode, "for node", nodeName)
		return reconcile.Result{}, err
	}

	if fencingMode != "delete" && fencingMode != "none" {
		// Setting new condition
		var newConditions []v1.NodeCondition

		for _, c := range node.Status.Conditions {
			if c.Type == v1.NodeReady || c.Reason == "NodeStatusUnknown" {
				c.Reason = "NodeFenced"
				c.Message = "Node was fenced by fencing controller."
				//TODO update time
			}
			newConditions = append(newConditions, c)
		}

		node.Status.Conditions = newConditions
		node.Status.VolumesAttached = nil
		node.Status.VolumesInUse = nil

		klog.Infoln("Updating node status", nodeName)
		err = r.client.Status().Update(context.Background(), node)
		if err != nil {
			klog.Error("Failed to patch node", nodeName, ":", err)
			return reconcile.Result{}, err
		}

		// Setting fencing status annotation
		mergePatch, _ := json.Marshal(map[string]interface{}{
			"metadata": map[string]interface{}{
				"annotations": map[string]interface{}{
					"fencing/state":     "fenced",
					"fencing/timestamp": nil,
				},
			},
		})
		err = r.client.Patch(context.TODO(), node, client.RawPatch(types.MergePatchType, mergePatch))
		if err != nil {
			klog.Errorln("Failed to patch node", node.Name, ":", err)
			return reconcile.Result{}, err
		}
		err = r.client.Patch(context.TODO(), instance, client.RawPatch(types.MergePatchType, mergePatch))
		if err != nil {
			klog.Errorln("Failed to patch job", instance.Name, ":", err)
			return reconcile.Result{}, err
		}
	}

	// Get after-hook annotation
	afterHook, ok := instance.Annotations["fencing/after-hook"]
	if !ok || afterHook == "" {
		return reconcile.Result{}, nil
	}
	klog.Infoln("Executing after hook", afterHook, "for", node.Name)

	// Find PodTemplate for after hook
	podTemplate := &v1.PodTemplate{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: afterHook, Namespace: instance.Namespace}, podTemplate)
	if err != nil && errors.IsNotFound(err) {
		klog.Errorln("Failed to find podTemplate ", afterHook, ":", err)
		return reconcile.Result{}, nil
	}

	// Define a new Job object
	job := newJobForJob(instance, podTemplate)

	// Check if this Job already exists
	found := &batchv1.Job{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: job.Name, Namespace: job.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		klog.Infoln("Creating a new job", job.Name)
		err = r.client.Create(context.TODO(), job)
		if err != nil {
			klog.Errorln("Failed to create new Job", ":", err)
			return reconcile.Result{}, err
		}

		// Job created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Job already exists - don't requeue
	klog.Infoln("Skip reconcile: job already exists", found.Name)
	return reconcile.Result{}, nil
}

// newJobForJob returns a job with afterHook for the fencing job
func newJobForJob(job *batchv1.Job, podTemplate *v1.PodTemplate) *batchv1.Job {
	labels := map[string]string{
		"node":    job.Annotations["fencing/node"],
		"fencing": "after-hook",
	}
	// Default annotations
	annotations := map[string]string{
		"fencing/mode":       job.Annotations["fencing/mode"],
		"fencing/template":   job.Annotations["fencing/template"],
		"fencing/after-hook": job.Annotations["fencing/after-hook"],
		"fencing/node":       job.Annotations["fencing/node"],
		"fencing/id":         job.Annotations["fencing/id"],
	}

	// Create new pod from podTemplate
	pod := podTemplate.Template
	// Apply annotations to the pod
	pod.ObjectMeta.Annotations = annotations

	// Set prefix name
	suffix := pod.Name
	if suffix == "" {
		suffix = "after-hook"
	}

	// Creating new Job
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:        job.Name + "-" + suffix,
			Namespace:   job.Namespace,
			Labels:      labels,
			Annotations: annotations,
			OwnerReferences: []metav1.OwnerReference{
				metav1.OwnerReference{
					APIVersion: job.APIVersion,
					Kind:       job.Kind,
					Name:       job.Name,
					UID:        job.UID,
				},
			},
		},
		Spec: batchv1.JobSpec{
			Template: pod},
	}
}
