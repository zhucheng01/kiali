package virtualservices

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kiali/kiali/kubernetes"
	"github.com/kiali/kiali/models"
	"github.com/kiali/kiali/tests/data"
	"github.com/kiali/kiali/tests/testutils/validations"
)

// VirtualService has two routes that all the weights sum 100
func TestServiceWellVirtualServiceValidation(t *testing.T) {
	assert := assert.New(t)

	// Setup mocks
	vals, valid := RouteChecker{fakeValidVirtualService()}.Check()

	// Well configured object
	assert.True(valid)
	assert.Empty(vals)
}

// VirtualService with one route and a weight between 0 and 100
func TestServiceMultipleChecks(t *testing.T) {
	assert := assert.New(t)

	vals, valid := RouteChecker{fakeOneRouteUnder100()}.Check()

	// wrong weight'ed route rule
	assert.True(valid)
	assert.NotEmpty(vals)
	assert.Len(vals, 1)
	assert.NoError(validations.ConfirmIstioCheckMessage("virtualservices.route.singleweight", vals[0]))
	assert.Equal(vals[0].Severity, models.WarningSeverity)
	assert.Equal(vals[0].Path, "spec/http[0]/route[0]/weight")
}

func TestVSWithRepeatingSubsets(t *testing.T) {
	assert := assert.New(t)

	vals, valid := RouteChecker{fakeRepeatedSubset()}.Check()
	assert.True(valid)
	assert.NotEmpty(vals)
	assert.Len(vals, 4)
	assert.NoError(validations.ConfirmIstioCheckMessage("virtualservices.route.repeatedsubset", vals[0]))
	assert.Equal(vals[0].Severity, models.WarningSeverity)
	assert.Regexp(`spec\/http\[0\]\/route\[[0,2]\]\/subset`, vals[0].Path)
	assert.NoError(validations.ConfirmIstioCheckMessage("virtualservices.route.repeatedsubset", vals[3]))
	assert.Equal(vals[3].Severity, models.WarningSeverity)
	assert.Regexp(`spec\/http\[0\]\/route\[[1,3]\]\/subset`, vals[3].Path)
}

func fakeValidVirtualService() kubernetes.IstioObject {
	validVirtualService := data.AddRoutesToVirtualService("http", data.CreateRoute("reviews", "v1", 55),
		data.AddRoutesToVirtualService("http", data.CreateRoute("reviews", "v2", 45),
			data.CreateEmptyVirtualService("reviews-well", "test", []string{"reviews"}),
		),
	)

	return validVirtualService
}

func fakeOneRouteUnder100() kubernetes.IstioObject {
	virtualService := data.AddRoutesToVirtualService("http", data.CreateRoute("reviews", "v1", 45),
		data.CreateEmptyVirtualService("reviews-multiple", "test", []string{"reviews"}),
	)

	return virtualService
}

func fakeRepeatedSubset() kubernetes.IstioObject {
	validVirtualService := data.AddRoutesToVirtualService("http", data.CreateRoute("reviews", "v1", 55),
		data.AddRoutesToVirtualService("http", data.CreateRoute("reviews", "v1", 45),
			data.AddRoutesToVirtualService("http", data.CreateRoute("reviews", "v2", 55),
				data.AddRoutesToVirtualService("http", data.CreateRoute("reviews", "v2", 45),
					data.CreateEmptyVirtualService("reviews-repeated", "test", []string{"reviews"}),
				),
			),
		),
	)

	return validVirtualService
}
