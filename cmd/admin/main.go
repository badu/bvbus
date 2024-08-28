package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/badu/bvbus/app/admin"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func GETBussesList(logger *slog.Logger, repo *admin.Repository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		busses, err := repo.GetBusses(len(r.URL.Query().Get("notCrawled")) > 0)
		if err != nil {
			logger.Error("error querying bus", "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(busses)
		if err != nil {
			logger.Error("error encoding json", "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusInternalServerError)
			return
		}
	}
}

func GETBusStations(logger *slog.Logger, repo *admin.Repository) func(w http.ResponseWriter, r *http.Request) {
	type BusWithAliasses struct {
		admin.Busline
		Aliases []admin.Alias `json:"a"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		busID := r.PathValue("busId")
		busId, err := strconv.ParseInt(busID, 10, 64)
		if err != nil {
			logger.Error("error parsing station id", "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusBadRequest)
			return
		}

		bus, err := repo.GetFullStationsBusByID(busId)
		if err != nil {
			logger.Error("error getting bus with stations", "id", busID, "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusInternalServerError)
			return
		}

		b, _ := json.MarshalIndent(bus, "\t", "")
		logger.Info("bus", "data", string(b))

		result := BusWithAliasses{Busline: *bus}
		result.Aliases, err = admin.CrawlStationNamesAndLinks(logger, bus)
		if err != nil {
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(result)
		if err != nil {
			logger.Error("error encoding json", "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusInternalServerError)
			return
		}
	}
}

func GETCrawl(logger *slog.Logger, repo *admin.Repository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		busID := r.PathValue("id")
		id, err := strconv.ParseInt(busID, 10, 64)
		if err != nil {
			logger.Error("error parsing bus id", "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusBadRequest)
			return
		}

		bus, err := repo.GetBusByID(id)
		if err != nil {
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusInternalServerError)
			return
		}

		aliases, err := admin.CrawlStationNamesAndLinks(logger, bus)
		if err != nil {
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(aliases)
		if err != nil {
			logger.Error("error encoding response", "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusInternalServerError)
			return
		}
	}
}

func GETBusWithStations(logger *slog.Logger, repo *admin.Repository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		busID := r.PathValue("id")
		id, err := strconv.ParseInt(busID, 10, 64)
		if err != nil {
			logger.Error("error parsing bus id", "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusBadRequest)
			return
		}

		bus, err := repo.GetFullStationsBusByID(id)
		if err != nil {
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(bus)
		if err != nil {
			logger.Error("error encoding response", "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusInternalServerError)
			return
		}
	}
}

func GETSave(logger *slog.Logger, repo *admin.Repository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		busID := r.PathValue("id")
		id, err := strconv.ParseInt(busID, 10, 64)
		if err != nil {
			logger.Error("error parsing bus id", "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusBadRequest)
			return
		}

		bus, err := repo.GetBusByID(id)
		if err != nil {
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusInternalServerError)
			return
		}

		logger.Info("start of saving crawled time tables", "id", id)

		timetablesMap, err := admin.CrawlTimeTables(logger, bus)
		if err != nil {
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusInternalServerError)
			return
		}

		err = repo.SaveTimeTables(bus.OSMID, timetablesMap)
		if err != nil {
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("{%q:%q}", "success", "true")))
	}
}

func GETStationTimetable(logger *slog.Logger, repo *admin.Repository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		stationID := r.PathValue("id")
		id, err := strconv.ParseInt(stationID, 10, 64)
		if err != nil {
			logger.Error("error parsing station id", "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusBadRequest)
			return
		}

		busID := r.PathValue("busId")
		busId, err := strconv.ParseInt(busID, 10, 64)
		if err != nil {
			logger.Error("error parsing station id", "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusInternalServerError)
			return
		}

		times, err := repo.GetTimeTablesForStationAndBus(id, busId)
		if err != nil {
			logger.Error("error querying timetables", "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(times)
		if err != nil {
			logger.Error("error encoding response", "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusInternalServerError)
			return
		}

	}
}

func GETAllStations(logger *slog.Logger, repo *admin.Repository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		stations, err := repo.GetStations()
		if err != nil {
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(stations)
		if err != nil {
			logger.Error("error encoding response", "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusInternalServerError)
			return
		}
	}
}

func GETQueryStreets(logger *slog.Logger, repo *admin.Repository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		bussesQuery := "data=[out:json];(relation[\"network\"=\"RAT Brașov\"];);out body;>;out skel qt;"
		bussesResponse, err := http.Post("https://overpass-api.de/api/interpreter", "text/plain", strings.NewReader(bussesQuery))
		if err != nil {
			logger.Error("Error fetching data from Overpass API", "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusInternalServerError)
			return
		}
		defer bussesResponse.Body.Close()

		busses, err := io.ReadAll(bussesResponse.Body)
		if err != nil {
			logger.Error("Error reading response from Overpass API", "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusInternalServerError)
			return
		}

		w.Write(busses)
	}
}

func GETQueryCrossings(logger *slog.Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		crossingsQuery := "[out:json];\nnode[\"highway\"=\"crossing\"](45.615231,25.486965,45.693884,25.705404);\nout body;\n>;\nout skel qt;"
		crossingsResponse, err := http.Post("https://overpass-api.de/api/interpreter", "text/plain", strings.NewReader(crossingsQuery))
		if err != nil {
			logger.Error("Error fetching data from Overpass API", "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusInternalServerError)
			return
		}
		defer crossingsResponse.Body.Close()

		crossings, err := io.ReadAll(crossingsResponse.Body)
		if err != nil {
			logger.Error("Error reading response from Overpass API", "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusInternalServerError)
			return
		}

		var crossingsData admin.Data
		err = json.Unmarshal(crossings, &crossingsData)
		if err != nil {
			logger.Error("Error parsing JSON", "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusInternalServerError)
			return
		}

		result := make([]admin.Node, 0)
		for _, crossing := range crossingsData.Elements {
			result = append(result, admin.Node{ID: crossing.ID, Lat: crossing.Lat, Lon: crossing.Lon})
		}

		responseData, err := json.Marshal(result)
		if err != nil {
			logger.Error("Error marshaling response JSON", "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusInternalServerError)
			return
		}

		w.Write(responseData)
	}
}

func POSTCrossing(logger *slog.Logger, repo *admin.Repository) func(w http.ResponseWriter, r *http.Request) {
	type Payload struct {
		ID         int64   `json:"id"`
		Lat        float64 `json:"lat"`
		Lon        float64 `json:"lon"`
		Station1ID int64   `json:"s1Id"`
		Station2ID int64   `json:"s2Id"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var payload Payload

		err := json.NewDecoder(r.Body).Decode(&payload)
		if err != nil {
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusBadRequest)
			logger.Error("error decoding request body", "err", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}
}

func GETStreetAddress(logger *slog.Logger, repo *admin.Repository) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		stationID := r.PathValue("id")
		id, err := strconv.ParseInt(stationID, 10, 64)
		if err != nil {
			logger.Error("error parsing station id", "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusBadRequest)
			return
		}

		station, err := repo.GetStationByID(id)
		if err != nil {
			logger.Error("error retrieving station by id", "id", stationID, "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusBadRequest)
			return
		}

		streetName, err := admin.ReverseGeocodeStreet(station.Lat, station.Lon)
		if err != nil {
			logger.Error("error retrieving street name", "id", stationID, "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusBadRequest)
			return
		}

		if streetName == nil {
			logger.Error("street name is nil", "id", stationID, "err", err)
			http.Error(w, fmt.Sprintf("{%q:%q}", "error", "street name is nil"), http.StatusBadRequest)
			return

		}

		if len(r.URL.Query().Get("save")) > 0 {
			err = repo.UpdateStreetName(id, *streetName)
			if err != nil {
				logger.Error("error updating street name", "id", stationID, "err", err)
				http.Error(w, fmt.Sprintf("{%q:%q}", "error", err.Error()), http.StatusBadRequest)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"street": *streetName})
	}
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	repo, err := admin.NewRepository(logger, "./data/brasov_busses.sqlite")
	if err != nil {
		logger.Error("error creating repository", "err", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())

	mux := http.NewServeMux()

	mux.HandleFunc("GET /lines", GETBussesList(logger, repo))
	mux.HandleFunc("GET /stations", GETAllStations(logger, repo))
	mux.HandleFunc("GET /stations/{busId}", GETBusStations(logger, repo))
	mux.HandleFunc("GET /station/{id}/{busId}", GETStationTimetable(logger, repo))
	mux.HandleFunc("GET /bus/{id}", GETBusWithStations(logger, repo))
	mux.HandleFunc("GET /crawl/{id}", GETCrawl(logger, repo))
	mux.HandleFunc("GET /save/{id}", GETSave(logger, repo))
	mux.HandleFunc("GET /streets/load", GETQueryStreets(logger, repo))
	mux.HandleFunc("GET /crossings", GETQueryCrossings(logger))
	mux.HandleFunc("POST /crossings", POSTCrossing(logger, repo))
	mux.HandleFunc("GET /tiles/{z}/{x}/{y}", admin.ServeTiles(logger, "./data/brasov.osm.pbf", repo))
	mux.HandleFunc("GET /revgeo/{id}", GETStreetAddress(logger, repo))

	go func() {
		logger.Info("Server started")
		server := &http.Server{Addr: ":8080", Handler: corsMiddleware(mux)}
		if err := server.ListenAndServe(); err != nil {
			logger.Error("error starting server", "err", err)
		}
	}()

	stoppedCh := make(chan os.Signal, 1)
	defer close(stoppedCh)
	signal.Notify(stoppedCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		logger.Info("waiting for stop signal...")
		<-stoppedCh
		logger.Info("received termination signal. cancelling context")
		cancel()
	}()

	<-ctx.Done()
	logger.Info("shutting down")
}
