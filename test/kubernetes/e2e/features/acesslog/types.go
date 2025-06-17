package acesslog

import (
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1alpha1 "github.com/kgateway-dev/kgateway/v2/api/v1alpha1"
	"github.com/kgateway-dev/kgateway/v2/pkg/utils/fsutils"
	e2edefaults "github.com/kgateway-dev/kgateway/v2/test/kubernetes/e2e/defaults"
	"github.com/kgateway-dev/kgateway/v2/test/kubernetes/e2e/tests/base"
)

var (
	// manifests
	setupManifest       = filepath.Join(fsutils.MustGetThisDir(), "testdata", "setup.yaml")
	fileSinkManifest    = filepath.Join(fsutils.MustGetThisDir(), "testdata", "filesink.yaml")
	grpcServiceManifest = filepath.Join(fsutils.MustGetThisDir(), "testdata", "grpc.yaml")
	OTelManifest        = filepath.Join(fsutils.MustGetThisDir(), "testdata", "otel.yaml")

	// Core infrastructure objects that we need to track
	gatewayObjectMeta = metav1.ObjectMeta{
		Name:      "gw",
		Namespace: "default",
	}
	gatewayService    = &corev1.Service{ObjectMeta: gatewayObjectMeta}
	gatewayDeployment = &appsv1.Deployment{ObjectMeta: gatewayObjectMeta}

	httpbinObjectMeta = metav1.ObjectMeta{
		Name:      "httpbin",
		Namespace: "httpbin",
	}
	httpbinDeployment = &appsv1.Deployment{ObjectMeta: httpbinObjectMeta}

	// TestAccessLogWithFileSink
	fileSinkConfig = &v1alpha1.HTTPListenerPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "access-logs",
			Namespace: "default",
		},
	}

	// TestAccessLogWithGrpcSink
	accessLoggerObjectMeta = metav1.ObjectMeta{
		Name:      "gateway-proxy-access-logger",
		Namespace: "default",
	}
	accessLoggerDeployment = &appsv1.Deployment{ObjectMeta: accessLoggerObjectMeta}
	accessLoggerService    = &corev1.Service{ObjectMeta: accessLoggerObjectMeta}

	// TestAccessLogWithOTelSink
	otelCollectorDeployment = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "otlpeek",
			Namespace: "default",
		},
	}

	setup = base.SimpleTestCase{
		Manifests: []string{e2edefaults.CurlPodManifest, setupManifest},
		Resources: []client.Object{e2edefaults.CurlPod, httpbinDeployment, gatewayService, gatewayDeployment},
	}

	// test cases
	testCases = map[string]*base.TestCase{
		"TestAccessLogWithFileSink": {
			SimpleTestCase: base.SimpleTestCase{
				Manifests: []string{fileSinkManifest},
				Resources: []client.Object{fileSinkConfig},
			},
		},
		"TestAccessLogWithGrpcSink": {
			SimpleTestCase: base.SimpleTestCase{
				Manifests: []string{grpcServiceManifest},
				Resources: []client.Object{accessLoggerService, accessLoggerDeployment},
			},
		},
		"TestAccessLogWithOTelSink": {
			SimpleTestCase: base.SimpleTestCase{
				Manifests: []string{OTelManifest},
				Resources: []client.Object{otelCollectorDeployment},
			},
		},
	}
)
