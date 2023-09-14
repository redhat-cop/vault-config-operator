/*
	Logging capable event handler that mimics handler.EnqueueRequestForObject
	See "sigs.k8s.io/controller-runtime/pkg/handler"
*/

package k8sevt

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
)

// List all types which include unexported fields so that cmp.Diff won't choke on them
var ignoredUnexportedDuringDiff = cmpopts.IgnoreUnexported(
	redhatcopv1alpha1.VRole{},
	redhatcopv1alpha1.DBSEConfig{},
	redhatcopv1alpha1.GHConfig{},
	redhatcopv1alpha1.JWTOIDCConfig{},
	redhatcopv1alpha1.KAECConfig{},
	redhatcopv1alpha1.KubeSEConfig{},
	redhatcopv1alpha1.LDAPConfig{},
	redhatcopv1alpha1.PKIIntermediate{},
	redhatcopv1alpha1.QuayConfig{},
	redhatcopv1alpha1.RMQSEConfig{},
	redhatcopv1alpha1.RandomSecretSpec{},
	redhatcopv1alpha1.GroupAliasSpec{},
)

var handlerLog = ctrl.Log.WithName("eventhandler")

type Log struct {
	predicate.Funcs
}

func (Log) Update(evt event.UpdateEvent) bool {
	return LogEventWithDiff("UpdateEvent", evt.ObjectOld, evt.ObjectNew)
}

func (Log) Create(evt event.CreateEvent) bool {
	return LogEvent("CreateEvent", evt.Object, evt)
}

func (Log) Delete(evt event.DeleteEvent) bool {
	return LogEvent("DeleteEvent", evt.Object, evt)
}

func (Log) Generic(evt event.GenericEvent) bool {
	return LogEvent("GenericEvent", evt.Object, evt)
}

func LogEvent(eventName string, object client.Object, evt interface{}) bool {
	handlerLog.V(1).Info(eventName+" received", "namespace", object.GetNamespace(), "name", object.GetName(), "event", evt)
	return true
}

func LogEventWithDiff(eventName string, objectOld client.Object, objectNew client.Object) bool {
	if handlerLog.V(1).Enabled() {
		switch {
		case objectNew != nil:
			handlerLog.V(1).Info(eventName+" received", "namespace", objectNew.GetNamespace(), "name", objectNew.GetName(),
				"diff", cmp.Diff(objectOld, objectNew, ignoredUnexportedDuringDiff))
		case objectOld != nil:
			handlerLog.V(1).Info(eventName+" received", "namespace", objectNew.GetNamespace(), "name", objectNew.GetName(),
				"diff", cmp.Diff(objectOld, objectNew, ignoredUnexportedDuringDiff))
		}
	}
	return true
}
