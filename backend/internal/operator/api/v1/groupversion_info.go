// Package v1 contains CollectorAgent API types.
package v1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

const (
	Group   = "observability.neuralops.io"
	Version = "v1"
)

var (
	GroupVersion = schema.GroupVersion{Group: Group, Version: Version}
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}
	AddToScheme = SchemeBuilder.AddToScheme
)
