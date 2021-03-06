package main

import (
	"lsh-search-service/app"
	cm "lsh-search-service/common"
	"net/http"
)

func main() {
	logger := cm.GetNewLogger()
	config, err := app.ParseEnv()
	if err != nil {
		logger.Err.Fatal(err.Error())
	}
	annServer, err := app.NewANNServer(logger, config)
	if err != nil {
		logger.Err.Fatal(err.Error())
	}
	defer annServer.Mongo.Disconnect()

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.HealthCheck)
	mux.HandleFunc("/build-index", annServer.BuildHasherHandler)
	mux.HandleFunc("/check-build", annServer.CheckBuildHandler)
	mux.HandleFunc("/get-nn", annServer.GetNeighborsHandler)
	mux.HandleFunc("/get-index-size", annServer.GetHashCollSizeHandler)
	mux.HandleFunc("/pop-hash", annServer.PopHashRecordHandler)
	mux.HandleFunc("/put-hash", annServer.PutHashRecordHandler)
	http.Handle("/", cm.Decorate(mux, cm.Timer(logger)))
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Err.Fatalf("Error running the server: %v", err)
	}
}
