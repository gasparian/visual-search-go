package lsh_test

import (
	"gonum.org/v1/gonum/blas/blas64"
	cm "lsh-search-service/common"
	hashing "lsh-search-service/lsh"
	"math"
	"testing"
)

func TestGetHash(t *testing.T) {
	hasherInstance := hashing.HasherInstance{
		Planes: []hashing.Plane{
			hashing.Plane{
				Coefs: cm.NewVec([]float64{1.0, 1.0, 1.0}),
				D:     5,
			},
		},
	}
	inpVec := cm.NewVec([]float64{5.0, 1.0, 1.0})
	meanVec := cm.NewVec([]float64{0.0, 0.0, 0.0})
	hash := hasherInstance.GetHash(inpVec, meanVec)
	if hash != 1 {
		t.Fatal("Wrong hash value, must be 1")
	}
	inpVec = cm.NewVec([]float64{1.0, 1.0, 1.0})
	hash = hasherInstance.GetHash(inpVec, meanVec)
	if hash != 0 {
		t.Fatal("Wrong hash value, must be 0")
	}
}

func getNewHasher(config hashing.Config) (*hashing.Hasher, error) {
	hasher := hashing.NewLSHIndex(config)
	mean := cm.NewVec([]float64{0.0, 0.0, 0.0})
	std := cm.NewVec([]float64{0.2, 0.3, 0.15})
	err := hasher.Generate(mean, std)
	if err != nil {
		return nil, err
	}
	return hasher, nil
}

func TestGenerateAngular(t *testing.T) {
	config := hashing.Config{
		IsAngularDistance: 1,
		NPermutes:         2,
		NPlanes:           1,
		BiasMultiplier:    2.0,
		DistanceThrsh:     0.8,
		Dims:              3,
		Bias:              4.0,
	}
	hasherAngular, err := getNewHasher(config)
	if err != nil {
		t.Fatalf("Smth went wrong with planes generation: %v", err)
	}

	isHasherEmpty := cm.IsZeroVector(hasherAngular.Instances[0].Planes[0].Coefs) ||
		cm.IsZeroVector(hasherAngular.Instances[0].Planes[0].Coefs)
	if isHasherEmpty {
		t.Fatal("One of the hasher instances is empty")
	}
}
func TestGenerateL2(t *testing.T) {
	config := hashing.Config{
		IsAngularDistance: 0,
		NPermutes:         2,
		NPlanes:           1,
		BiasMultiplier:    2.0,
		DistanceThrsh:     0.8,
		Dims:              3,
		Bias:              4.0,
	}
	hasher, err := getNewHasher(config)
	if err != nil {
		t.Fatalf("Smth went wrong with planes generation: %v", err)
	}
	var distToOrigin float64
	maxDist := config.Bias * config.BiasMultiplier
	for _, hasherInstance := range hasher.Instances {
		for _, plane := range hasherInstance.Planes {
			distToOrigin = math.Abs(plane.D) / blas64.Nrm2(plane.Coefs)
			if distToOrigin > maxDist {
				t.Fatal("Generated plane is out of bounds defined by hasher config")
			}
		}
	}
}

func TestGetHashes(t *testing.T) {
	config := hashing.Config{
		IsAngularDistance: 1,
		NPermutes:         2,
		NPlanes:           1,
		BiasMultiplier:    2.0,
		DistanceThrsh:     0.8,
		Dims:              3,
		Bias:              4.0,
	}
	hasherAngular, err := getNewHasher(config)
	if err != nil {
		t.Fatalf("Smth went wrong with planes generation: %v", err)
	}
	inpVec := cm.NewVec([]float64{0.0, 0.0, 0.0})
	hashes := hasherAngular.GetHashes(inpVec)
	for _, v := range hashes {
		if v != 1 {
			t.Fatal("Hash should always be 1 at this case")
		}
	}
}

func TestGetDistAngular(t *testing.T) {
	config := hashing.Config{
		IsAngularDistance: 1,
		NPermutes:         2,
		NPlanes:           1,
		BiasMultiplier:    2.0,
		DistanceThrsh:     0.8,
		Dims:              3,
		Bias:              4.0,
	}
	hasherAngular, err := getNewHasher(config)
	if err != nil {
		t.Fatalf("Smth went wrong with planes generation: %v", err)
	}
	v1 := cm.NewVec([]float64{0.0, 0.0, 0.0})
	v2 := cm.NewVec([]float64{0.0, 1.0, 0.0})
	dist, ok := hasherAngular.GetDist(v1, v2)
	if ok {
		t.Fatal("Angular distance can't be calculated properly with zero vector")
	}
	v1 = cm.NewVec([]float64{0.0, 0.0, 2.0})
	v2 = cm.NewVec([]float64{0.0, 1.0, 0.0})
	dist, _ = hasherAngular.GetDist(v1, v2)
	if ok {
		t.Fatal("Measured dist must be greater than the threshold")
	}
	if dist != 1.0 {
		t.Fatal("Measured dist must be equal to 1.0")
	}
}

func TestGetDistL2(t *testing.T) {
	config := hashing.Config{
		IsAngularDistance: 0,
		NPermutes:         2,
		NPlanes:           1,
		BiasMultiplier:    2.0,
		DistanceThrsh:     1.1,
		Dims:              3,
		Bias:              4.0,
	}
	hasher, err := getNewHasher(config)
	if err != nil {
		t.Fatalf("Smth went wrong with planes generation: %v", err)
	}
	v1 := cm.NewVec([]float64{0.0, 0.0, 0.0})
	v2 := cm.NewVec([]float64{0.0, 1.0, 0.0})
	dist, ok := hasher.GetDist(v1, v2)
	if !ok {
		t.Fatal("L2 distance must pass the threshold")
	}
	if dist != 1.0 {
		t.Fatal("L2 distance must be equal to 1.0")
	}
}

func TestDumpHasher(t *testing.T) {
	config := hashing.Config{
		IsAngularDistance: 0,
		NPermutes:         2,
		NPlanes:           1,
		BiasMultiplier:    2.0,
		DistanceThrsh:     1.1,
		Dims:              3,
		Bias:              4.0,
	}
	hasher, err := getNewHasher(config)
	if err != nil {
		t.Fatalf("Smth went wrong with planes generation: %v", err)
	}
	coefToTest := hasher.Instances[0].Planes[0].D
	b, err := hasher.Dump()
	if err != nil {
		t.Fatalf("Could not serialize hasher: %v", err)
	}
	if len(b) == 0 {
		t.Fatal("Smth went wrong serializing the hasher: resulting bytearray is empty")
	}

	err = hasher.Load(b)
	if err != nil {
		t.Fatalf("Could not deserialize hasher: %v", err)
	}
	if coefToTest != hasher.Instances[0].Planes[0].D {
		t.Fatal("Seems like the deserialized hasher differs from the initial one")
	}
}
