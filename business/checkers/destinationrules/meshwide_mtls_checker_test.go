package destinationrules

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/models"
	"github.com/kiali/kiali/tests/data"
	"github.com/kiali/kiali/tests/testutils/validations"
)

// Context: DestinationRule enables mesh-wide mTLS
// Context: There is no MeshPolicy
// It returns any validation
func TestMTLSMeshWideDREnabledWithNoMeshPolicy(t *testing.T) {
	destinationRule := data.AddTrafficPolicyToDestinationRule(data.CreateMTLSTrafficPolicyForDestinationRules(),
		data.CreateEmptyDestinationRule("istio-system", "dr-mtls", "*.local"))

	mTlsDetails := kubernetes.MTLSDetails{
		MeshPeerAuthentications: []kubernetes.IstioObject{},
	}

	testReturnsAValidation(t, destinationRule, mTlsDetails)
}

// Context: DestinationRule enables mesh-wide mTLS
// Context: There is one MeshPolicy in PERMISSIVE mode
// It doesn't return any validation
func TestMTLSMeshWideDREnabledWithMeshPolicyDisabled(t *testing.T) {
	destinationRule := data.AddTrafficPolicyToDestinationRule(data.CreateMTLSTrafficPolicyForDestinationRules(),
		data.CreateEmptyDestinationRule("istio-system", "dr-mtls", "*.local"))

	mTlsDetails := kubernetes.MTLSDetails{
		MeshPeerAuthentications: []kubernetes.IstioObject{
			data.CreateEmptyMeshPeerAuthentication("default", data.CreateMTLS("PERMISSIVE")),
		},
	}

	testNoValidationsFound(t, destinationRule, mTlsDetails)
}

// Context: DestinationRule enables mesh-wide mTLS
// Context: There is one MeshPolicy enabling mTLS in STRICT mode
// It doesn't return any validation
func TestMTLSMeshWideDREnabledWithMeshPolicy(t *testing.T) {
	destinationRule := data.AddTrafficPolicyToDestinationRule(data.CreateMTLSTrafficPolicyForDestinationRules(),
		data.CreateEmptyDestinationRule("istio-system", "dr-mtls", "*.local"))

	mTlsDetails := kubernetes.MTLSDetails{
		MeshPeerAuthentications: []kubernetes.IstioObject{
			data.CreateEmptyMeshPeerAuthentication("default", data.CreateMTLS("STRICT")),
		},
	}

	testNoValidationsFound(t, destinationRule, mTlsDetails)
}

// Context: DestinationRule enables namespace-wide mTLS
// Context: There is one MeshPolicy enabling mTLS in STRICT mode
// It doesn't return any validation
func TestMTLSNamespaceWideDREnabledWithMeshPolicy(t *testing.T) {
	destinationRule := data.AddTrafficPolicyToDestinationRule(data.CreateMTLSTrafficPolicyForDestinationRules(),
		data.CreateEmptyDestinationRule("istio-system", "dr-mtls", "*.istio-system.svc.cluster.local"))

	mTlsDetails := kubernetes.MTLSDetails{
		MeshPeerAuthentications: []kubernetes.IstioObject{
			data.CreateEmptyMeshPeerAuthentication("default", data.CreateMTLS("STRICT")),
		},
	}

	testNoValidationsFound(t, destinationRule, mTlsDetails)
}

// Context: DestinationRule enables namespace-wide mTLS
// Context: There is one MeshPolicy enabling mTLS in PERMISSIVE mode
// It doesn't return any validation
func TestMTLSNamespaceWideDREnabledWithMeshPolicyDisabled(t *testing.T) {
	destinationRule := data.AddTrafficPolicyToDestinationRule(data.CreateMTLSTrafficPolicyForDestinationRules(),
		data.CreateEmptyDestinationRule("istio-system", "dr-mtls", "*.istio-system.svc.cluster.local"))

	mTlsDetails := kubernetes.MTLSDetails{
		MeshPeerAuthentications: []kubernetes.IstioObject{
			data.CreateEmptyMeshPeerAuthentication("default", data.CreateMTLS("PERMISSIVE")),
		},
	}

	testNoValidationsFound(t, destinationRule, mTlsDetails)
}

// Context: DestinationRule not enabling mTLS
// Context: There is one MeshPolicy enabling mTLS
// It doesn't return any validation
func TestMTLSDRDisabledWithMeshPolicy(t *testing.T) {
	destinationRule := data.CreateEmptyDestinationRule("istio-system", "dr-mtls", "*.istio-system.svc.cluster.local")

	mTlsDetails := kubernetes.MTLSDetails{
		MeshPeerAuthentications: []kubernetes.IstioObject{
			data.CreateEmptyMeshPeerAuthentication("default", data.CreateMTLS("STRICT")),
		},
	}

	testNoValidationsFound(t, destinationRule, mTlsDetails)
}

// Context: DestinationRule not enabling mTLS
// Context: There is one MeshPolicy not enabling mTLS
// It doesn't return any validation
func TestMTLSDRDisabledWithMeshPolicyDisabled(t *testing.T) {
	destinationRule := data.CreateEmptyDestinationRule("istio-system", "dr-mtls", "*.istio-system.svc.cluster.local")

	mTlsDetails := kubernetes.MTLSDetails{
		MeshPeerAuthentications: []kubernetes.IstioObject{
			data.CreateEmptyMeshPeerAuthentication("default", data.CreateMTLS("PERMISSIVE")),
		},
	}

	testNoValidationsFound(t, destinationRule, mTlsDetails)
}

func testReturnsAValidation(t *testing.T, destinationRule kubernetes.IstioObject, mTLSDetails kubernetes.MTLSDetails) {
	assert := assert.New(t)

	vals, valid := MeshWideMTLSChecker{
		DestinationRule: destinationRule,
		MTLSDetails:     mTLSDetails,
	}.Check()

	assert.NotEmpty(vals)
	assert.Equal(1, len(vals))
	assert.False(valid)

	validation := vals[0]
	assert.NotNil(validation)
	assert.Equal(models.ErrorSeverity, validation.Severity)
	assert.Equal("spec/trafficPolicy/tls/mode", validation.Path)
	assert.NoError(validations.ConfirmIstioCheckMessage("destinationrules.mtls.meshpolicymissing", validation))
}

func testNoValidationsFound(t *testing.T, destinationRule kubernetes.IstioObject, mTLSDetails kubernetes.MTLSDetails) {
	assert := assert.New(t)

	validations, valid := MeshWideMTLSChecker{
		DestinationRule: destinationRule,
		MTLSDetails:     mTLSDetails,
	}.Check()

	assert.Empty(validations)
	assert.True(valid)
}
