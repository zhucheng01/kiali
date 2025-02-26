package virtualservices

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/models"
	"github.com/kiali/kiali/tests/data"
	"github.com/kiali/kiali/tests/testutils/validations"
)

func TestOneVirtualServicePerHost(t *testing.T) {
	vss := []kubernetes.IstioObject{
		buildVirtualService("virtual-1", "reviews"),
		buildVirtualService("virtual-2", "ratings"),
	}
	vals := SingleHostChecker{
		Namespace:               "bookinfo",
		VirtualServices:         vss,
		ExportedVirtualServices: []kubernetes.IstioObject{},
	}.Check()

	emptyValidationTest(t, vals)

	// First virtual service has a gateway
	vss = []kubernetes.IstioObject{
		buildVirtualServiceWithGateway("virtual-1", "reviews", "bookinfo-gateway"),
		buildVirtualService("virtual-2", "ratings"),
	}
	vals = SingleHostChecker{
		Namespace:               "bookinfo",
		VirtualServices:         vss,
		ExportedVirtualServices: []kubernetes.IstioObject{},
	}.Check()

	emptyValidationTest(t, vals)
	emptyValidationTest(t, vals)

	// Second virtual service has a gateway
	vss = []kubernetes.IstioObject{
		buildVirtualService("virtual-1", "reviews"),
		buildVirtualServiceWithGateway("virtual-2", "ratings", "bookinfo-gateway"),
	}
	vals = SingleHostChecker{
		Namespace:               "bookinfo",
		VirtualServices:         vss,
		ExportedVirtualServices: []kubernetes.IstioObject{},
	}.Check()

	emptyValidationTest(t, vals)
	emptyValidationTest(t, vals)

	// Both virtual services have a gateway
	vss = []kubernetes.IstioObject{
		buildVirtualServiceWithGateway("virtual-1", "reviews", "bookinfo-gateway"),
		buildVirtualServiceWithGateway("virtual-2", "ratings", "bookinfo-gateway"),
	}

	vals = SingleHostChecker{
		Namespace:               "bookinfo",
		VirtualServices:         vss,
		ExportedVirtualServices: []kubernetes.IstioObject{},
	}.Check()

	emptyValidationTest(t, vals)
	emptyValidationTest(t, vals)
}

func TestOneVirtualServicePerFQDNHost(t *testing.T) {
	vss := []kubernetes.IstioObject{
		buildVirtualService("virtual-1", "reviews.bookinfo.svc.cluster.local"),
		buildVirtualService("virtual-2", "ratings.bookinfo.svc.cluster.local"),
	}
	vals := SingleHostChecker{
		Namespace:               "bookinfo",
		VirtualServices:         vss,
		ExportedVirtualServices: []kubernetes.IstioObject{},
	}.Check()

	emptyValidationTest(t, vals)
}

func TestOneVirtualServicePerFQDNWildcardHost(t *testing.T) {
	vss := []kubernetes.IstioObject{
		buildVirtualService("virtual-1", "*.bookinfo.svc.cluster.local"),
		buildVirtualService("virtual-2", "*.eshop.svc.cluster.local"),
	}
	vals := SingleHostChecker{
		Namespace:               "bookinfo",
		VirtualServices:         vss,
		ExportedVirtualServices: []kubernetes.IstioObject{},
	}.Check()

	emptyValidationTest(t, vals)
}

func TestRepeatingSimpleHost(t *testing.T) {
	vss := []kubernetes.IstioObject{
		buildVirtualService("virtual-1", "reviews"),
		buildVirtualService("virtual-2", "reviews"),
		buildVirtualService("virtual-3", "reviews"),
	}

	vals := SingleHostChecker{
		Namespace:               "bookinfo",
		VirtualServices:         vss,
		ExportedVirtualServices: []kubernetes.IstioObject{},
	}.Check()

	presentValidationTest(t, vals, "virtual-1")
	presentValidationTest(t, vals, "virtual-2")
	presentValidationTest(t, vals, "virtual-3")

	for _, validation := range vals {
		switch validation.Name {
		case "virtual-1":
			presentReferences(t, *validation, []string{"virtual-2", "virtual-3"})
		case "virtual-2":
			presentReferences(t, *validation, []string{"virtual-1", "virtual-3"})
		case "virtual-3":
			presentReferences(t, *validation, []string{"virtual-1", "virtual-2"})
		}
	}
}

func TestRepeatingSimpleHostWithGateway(t *testing.T) {
	vss := []kubernetes.IstioObject{
		buildVirtualServiceWithGateway("virtual-1", "reviews", "bookinfo"),
		buildVirtualService("virtual-2", "reviews"),
	}

	vals := SingleHostChecker{
		Namespace:               "bookinfo",
		VirtualServices:         vss,
		ExportedVirtualServices: []kubernetes.IstioObject{},
	}.Check()

	noObjectValidationTest(t, vals, "virtual-1")
	noObjectValidationTest(t, vals, "virtual-2")

	vss = []kubernetes.IstioObject{
		buildVirtualService("virtual-1", "reviews"),
		buildVirtualServiceWithGateway("virtual-2", "reviews", "bookinfo"),
	}

	vals = SingleHostChecker{
		Namespace:               "bookinfo",
		VirtualServices:         vss,
		ExportedVirtualServices: []kubernetes.IstioObject{},
	}.Check()

	noObjectValidationTest(t, vals, "virtual-1")
	noObjectValidationTest(t, vals, "virtual-2")

	vss = []kubernetes.IstioObject{
		buildVirtualServiceWithGateway("virtual-1", "reviews", "bookinfo"),
		buildVirtualServiceWithGateway("virtual-2", "reviews", "bookinfo"),
	}

	vals = SingleHostChecker{
		Namespace:               "bookinfo",
		VirtualServices:         vss,
		ExportedVirtualServices: []kubernetes.IstioObject{},
	}.Check()

	refKey := models.IstioValidationKey{ObjectType: "virtualservice", Namespace: "bookinfo", Name: "virtual-2"}
	presentValidationTest(t, vals, "virtual-1")
	presentReference(t, *(vals[refKey]), "virtual-1")

	refKey.Name = "virtual-2"
	presentValidationTest(t, vals, "virtual-2")
	presentReference(t, *(vals[refKey]), "virtual-1")
}

func TestRepeatingSVCNSHost(t *testing.T) {
	vss := []kubernetes.IstioObject{
		buildVirtualService("virtual-1", "reviews.bookinfo"),
		buildVirtualService("virtual-2", "reviews.bookinfo"),
	}
	vals := SingleHostChecker{
		Namespace: "bookinfo",
		Namespaces: models.Namespaces{
			{Name: "bookinfo"},
		},
		VirtualServices:         vss,
		ExportedVirtualServices: []kubernetes.IstioObject{},
	}.Check()

	presentValidationTest(t, vals, "virtual-1")
	presentValidationTest(t, vals, "virtual-2")

	vss = []kubernetes.IstioObject{
		buildVirtualService("virtual-1", "reviews"),
		buildVirtualService("virtual-2", "reviews.bookinfo"),
	}
	vals = SingleHostChecker{
		Namespace: "bookinfo",
		Namespaces: models.Namespaces{
			{Name: "bookinfo"},
		},
		VirtualServices:         vss,
		ExportedVirtualServices: []kubernetes.IstioObject{},
	}.Check()

	presentValidationTest(t, vals, "virtual-1")
	presentValidationTest(t, vals, "virtual-2")

	vss = []kubernetes.IstioObject{
		buildVirtualService("virtual-1", "reviews.bookinfo.svc.cluster.local"),
		buildVirtualService("virtual-2", "reviews.bookinfo"),
		buildVirtualServiceWithGateway("virtual-3", "reviews", "bookinfo-gateway-auto"),
	}
	vals = SingleHostChecker{
		Namespace: "bookinfo",
		Namespaces: models.Namespaces{
			{Name: "bookinfo"},
		},
		VirtualServices:         vss,
		ExportedVirtualServices: []kubernetes.IstioObject{},
	}.Check()

	presentValidationTest(t, vals, "virtual-1")
	presentValidationTest(t, vals, "virtual-2")

	vss = []kubernetes.IstioObject{
		buildVirtualService("virtual-1", "*.bookinfo.svc.cluster.local"),
		buildVirtualService("virtual-2", "reviews.bookinfo"),
	}
	vals = SingleHostChecker{
		Namespace: "bookinfo",
		Namespaces: models.Namespaces{
			{Name: "bookinfo"},
		},
		VirtualServices:         vss,
		ExportedVirtualServices: []kubernetes.IstioObject{},
	}.Check()

	presentValidationTest(t, vals, "virtual-1")
	presentValidationTest(t, vals, "virtual-2")

	vss = []kubernetes.IstioObject{
		buildVirtualService("virtual-1", "reviews"),
		buildVirtualService("virtual-2", "details.bookinfo"),
	}
	vals = SingleHostChecker{
		Namespace: "bookinfo",
		Namespaces: models.Namespaces{
			{Name: "bookinfo"},
		},
		VirtualServices:         vss,
		ExportedVirtualServices: []kubernetes.IstioObject{},
	}.Check()

	noObjectValidationTest(t, vals, "virtual-1")
	noObjectValidationTest(t, vals, "virtual-2")
	emptyValidationTest(t, vals)

	vss = []kubernetes.IstioObject{
		buildVirtualService("virtual-1", "reviews.bookinfo.svc.cluster.local"),
		buildVirtualService("virtual-2", "details.bookinfo"),
	}
	vals = SingleHostChecker{
		Namespace: "bookinfo",
		Namespaces: models.Namespaces{
			{Name: "bookinfo"},
		},
		VirtualServices:         vss,
		ExportedVirtualServices: []kubernetes.IstioObject{},
	}.Check()

	noObjectValidationTest(t, vals, "virtual-1")
	noObjectValidationTest(t, vals, "virtual-2")
	emptyValidationTest(t, vals)
}

func TestRepeatingFQDNHost(t *testing.T) {
	vss := []kubernetes.IstioObject{
		buildVirtualService("virtual-1", "reviews.bookinfo.svc.cluster.local"),
		buildVirtualService("virtual-2", "reviews.bookinfo.svc.cluster.local"),
		buildVirtualService("virtual-3", "reviews.bookinfo.svc.cluster.local"),
	}
	vals := SingleHostChecker{
		Namespace:               "bookinfo",
		VirtualServices:         vss,
		ExportedVirtualServices: []kubernetes.IstioObject{},
	}.Check()

	presentValidationTest(t, vals, "virtual-1")
	presentValidationTest(t, vals, "virtual-2")
	presentValidationTest(t, vals, "virtual-3")

	for _, validation := range vals {
		switch validation.Name {
		case "virtual-1":
			presentReferences(t, *validation, []string{"virtual-2", "virtual-3"})
		case "virtual-2":
			presentReferences(t, *validation, []string{"virtual-1", "virtual-3"})
		case "virtual-3":
			presentReferences(t, *validation, []string{"virtual-1", "virtual-2"})
		}
	}
}

func TestRepeatingFQDNWildcardHost(t *testing.T) {
	vss := []kubernetes.IstioObject{
		buildVirtualService("virtual-1", "*.bookinfo.svc.cluster.local"),
		buildVirtualService("virtual-2", "*.bookinfo.svc.cluster.local"),
		buildVirtualService("virtual-3", "*.bookinfo.svc.cluster.local"),
	}
	vals := SingleHostChecker{
		Namespace:               "bookinfo",
		VirtualServices:         vss,
		ExportedVirtualServices: []kubernetes.IstioObject{},
	}.Check()

	presentValidationTest(t, vals, "virtual-1")
	presentValidationTest(t, vals, "virtual-2")
	presentValidationTest(t, vals, "virtual-3")

	for _, validation := range vals {
		switch validation.Name {
		case "virtual-1":
			presentReferences(t, *validation, []string{"virtual-2", "virtual-3"})
		case "virtual-2":
			presentReferences(t, *validation, []string{"virtual-1", "virtual-3"})
		case "virtual-3":
			presentReferences(t, *validation, []string{"virtual-1", "virtual-2"})
		}
	}
}

func TestIncludedIntoWildCard(t *testing.T) {
	vss := []kubernetes.IstioObject{
		buildVirtualService("virtual-1", "*.bookinfo.svc.cluster.local"),
		buildVirtualService("virtual-2", "reviews.bookinfo.svc.cluster.local"),
		buildVirtualService("virtual-3", "reviews.bookinfo.svc.cluster.local"),
	}
	vals := SingleHostChecker{
		Namespace:               "bookinfo",
		VirtualServices:         vss,
		ExportedVirtualServices: []kubernetes.IstioObject{},
	}.Check()

	presentValidationTest(t, vals, "virtual-1")
	presentValidationTest(t, vals, "virtual-2")
	presentValidationTest(t, vals, "virtual-3")

	for _, validation := range vals {
		switch validation.Name {
		case "virtual-1":
			presentReferences(t, *validation, []string{"virtual-2", "virtual-3"})
		case "virtual-2":
			presentReferences(t, *validation, []string{"virtual-1", "virtual-3"})
		case "virtual-3":
			presentReferences(t, *validation, []string{"virtual-1", "virtual-2"})
		}
	}

	// Same test, with different order of appearance
	vss = []kubernetes.IstioObject{
		buildVirtualService("virtual-1", "reviews.bookinfo.svc.cluster.local"),
		buildVirtualService("virtual-2", "*.bookinfo.svc.cluster.local"),
		buildVirtualService("virtual-3", "reviews.bookinfo.svc.cluster.local"),
	}
	vals = SingleHostChecker{
		Namespace:               "bookinfo",
		VirtualServices:         vss,
		ExportedVirtualServices: []kubernetes.IstioObject{},
	}.Check()

	presentValidationTest(t, vals, "virtual-1")
	presentValidationTest(t, vals, "virtual-2")
	presentValidationTest(t, vals, "virtual-3")

	for _, validation := range vals {
		switch validation.Name {
		case "virtual-1":
			presentReferences(t, *validation, []string{"virtual-2", "virtual-3"})
		case "virtual-2":
			presentReferences(t, *validation, []string{"virtual-1", "virtual-3"})
		case "virtual-3":
			presentReferences(t, *validation, []string{"virtual-1", "virtual-2"})
		}
	}
}

func TestShortHostNameIncludedIntoWildCard(t *testing.T) {
	vss := []kubernetes.IstioObject{
		buildVirtualService("virtual-1", "*.bookinfo.svc.cluster.local"),
		buildVirtualService("virtual-2", "reviews"),
		buildVirtualService("virtual-3", "reviews"),
	}
	vals := SingleHostChecker{
		Namespace:               "bookinfo",
		VirtualServices:         vss,
		ExportedVirtualServices: []kubernetes.IstioObject{},
	}.Check()

	presentValidationTest(t, vals, "virtual-1")
	presentValidationTest(t, vals, "virtual-2")
	presentValidationTest(t, vals, "virtual-3")

	for _, validation := range vals {
		switch validation.Name {
		case "virtual-1":
			presentReferences(t, *validation, []string{"virtual-2", "virtual-3"})
		case "virtual-2":
			presentReferences(t, *validation, []string{"virtual-1", "virtual-3"})
		case "virtual-3":
			presentReferences(t, *validation, []string{"virtual-1", "virtual-2"})
		}
	}
}

func TestWildcardisMarkedInvalid(t *testing.T) {
	vss := []kubernetes.IstioObject{
		buildVirtualService("virtual-1", "*"),
		buildVirtualService("virtual-2", "reviews"),
		buildVirtualService("virtual-3", "reviews"),
	}
	vals := SingleHostChecker{
		Namespace:               "bookinfo",
		VirtualServices:         vss,
		ExportedVirtualServices: []kubernetes.IstioObject{},
	}.Check()

	presentValidationTest(t, vals, "virtual-1")
	presentValidationTest(t, vals, "virtual-2")
	presentValidationTest(t, vals, "virtual-3")

	for _, validation := range vals {
		switch validation.Name {
		case "virtual-1":
			presentReferences(t, *validation, []string{"virtual-2", "virtual-3"})
		case "virtual-2":
			presentReferences(t, *validation, []string{"virtual-1", "virtual-3"})
		case "virtual-3":
			presentReferences(t, *validation, []string{"virtual-1", "virtual-2"})
		}
	}
}

func TestMultipleHostsFailing(t *testing.T) {
	vss := []kubernetes.IstioObject{
		buildVirtualService("virtual-1", "reviews"),
		buildVirtualServiceMultipleHosts("virtual-2", []string{"reviews",
			"mongo.backup.svc.cluster.local", "mongo.staging.svc.cluster.local"}),
	}
	vals := SingleHostChecker{
		Namespace:               "bookinfo",
		VirtualServices:         vss,
		ExportedVirtualServices: []kubernetes.IstioObject{},
	}.Check()

	presentValidationTest(t, vals, "virtual-1")
	presentValidationTest(t, vals, "virtual-2")

	for _, validation := range vals {
		switch validation.Name {
		case "virtual-1":
			presentReference(t, *validation, "virtual-2")
		case "virtual-2":
			presentReference(t, *validation, "virtual-1")
		}
	}
}

func TestMultipleHostsPassing(t *testing.T) {
	vss := []kubernetes.IstioObject{
		buildVirtualService("virtual-1", "reviews"),
		buildVirtualServiceMultipleHosts("virtual-2", []string{"ratings",
			"mongo.backup.svc.cluster.local", "mongo.staging.svc.cluster.local"}),
	}
	vals := SingleHostChecker{
		Namespace:               "bookinfo",
		VirtualServices:         vss,
		ExportedVirtualServices: []kubernetes.IstioObject{},
	}.Check()

	emptyValidationTest(t, vals)
}

func buildVirtualService(name, host string) kubernetes.IstioObject {
	return buildVirtualServiceMultipleHosts(name, []string{host})
}

func buildVirtualServiceWithGateway(name, host, gateway string) kubernetes.IstioObject {
	return data.AddGatewaysToVirtualService([]string{gateway}, data.CreateEmptyVirtualService(name,
		"bookinfo", []string{host}))
}

func buildVirtualServiceMultipleHosts(name string, hosts []string) kubernetes.IstioObject {
	return data.CreateEmptyVirtualService(name, "bookinfo", hosts)
}

func emptyValidationTest(t *testing.T, vals models.IstioValidations) {
	assert := assert.New(t)
	assert.Empty(vals)

	validation, ok := vals[models.IstioValidationKey{ObjectType: "virtualservice", Namespace: "bookinfo", Name: "virtual-1"}]
	assert.False(ok)
	assert.Nil(validation)

	validation, ok = vals[models.IstioValidationKey{ObjectType: "virtualservice", Namespace: "bookinfo", Name: "virtual-2"}]
	assert.False(ok)
	assert.Nil(validation)
}

func noObjectValidationTest(t *testing.T, vals models.IstioValidations, name string) {
	assert := assert.New(t)

	validation, ok := vals[models.IstioValidationKey{ObjectType: "virtualservice", Namespace: "bookinfo", Name: name}]
	assert.False(ok)
	assert.Nil(validation)
}

func presentValidationTest(t *testing.T, vals models.IstioValidations, serviceName string) {
	assert := assert.New(t)
	assert.NotEmpty(vals)

	validation, ok := vals[models.IstioValidationKey{ObjectType: "virtualservice", Namespace: "bookinfo", Name: serviceName}]
	assert.True(ok)

	assert.True(validation.Valid)
	assert.NotEmpty(validation.Checks)
	assert.Equal(models.WarningSeverity, validation.Checks[0].Severity)
	assert.NoError(validations.ConfirmIstioCheckMessage("virtualservices.singlehost", validation.Checks[0]))
	assert.Equal("spec/hosts", validation.Checks[0].Path)
}

func presentReference(t *testing.T, validation models.IstioValidation, serviceName string) {
	assert := assert.New(t)
	refKey := models.IstioValidationKey{ObjectType: "virtualservice", Namespace: "bookinfo", Name: serviceName}

	assert.True(len(validation.References) > 0)
	assert.Contains(validation.References, refKey)
}

func presentReferences(t *testing.T, validation models.IstioValidation, serviceNames []string) {
	assert := assert.New(t)
	assert.True(len(validation.References) > 0)

	for _, sn := range serviceNames {
		refKey := models.IstioValidationKey{ObjectType: "virtualservice", Namespace: "bookinfo", Name: sn}
		assert.Contains(validation.References, refKey)
	}
}
