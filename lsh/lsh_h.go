package lsh

import (
	"sync"

	"gonum.org/v1/gonum/blas/blas64"
)

// Plane struct holds data needed to work with plane
type Plane struct {
	Coefs blas64.Vector
	D     float64
}

// HasherInstance holds data for local sensetive hashing algorithm
type HasherInstance struct {
	Planes []Plane
}

// Config holds all needed constants for creating the Hasher instance
type Config struct {
	IsAngularDistance int
	NPermutes         int
	NPlanes           int
	BiasMultiplier    float64
	DistanceThrsh     float64
	Dims              int
	Bias              float64
	MeanVec           blas64.Vector
}

// Hasher holds N_PERMUTS number of HasherInstance instances
type Hasher struct {
	sync.Mutex
	Config          Config
	Instances       []HasherInstance
	HashFieldsNames []string
}

// HasherEncode using for encoding/decoding the Hasher structure
type HasherEncode struct {
	Instances       *[]HasherInstance
	HashFieldsNames *[]string
	Config          *Config
}

// SafeHashesHolder allows to lock map while write values in it
type safeHashesHolder struct {
	sync.Mutex
	v map[int]uint64
}
