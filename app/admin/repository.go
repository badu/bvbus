package admin

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

const (
	ForeignKeysSQL = `PRAGMA foreign_keys=on;`

	CreateBussesSQL       = `CREATE TABLE IF NOT EXISTS busses (id INTEGER PRIMARY KEY,dir INTEGER DEFAULT 1,name TEXT NOT NULL,from_station TEXT NOT NULL,to_station TEXT NOT NULL,no TEXT NOT NULL,color TEXT,website TEXT,urban BOOLEAN DEFAULT 0,metropolitan BOOLEAN DEFAULT 0,crawled BOOLEAN DEFAULT 0);`
	CreateStationsSQL     = `CREATE TABLE IF NOT EXISTS stations (id INTEGER PRIMARY KEY,name TEXT NOT NULL,street_name TEXT NULL,lat REAL NOT NULL,lng REAL NOT NULL,outside BOOLEAN DEFAULT 0,board BOOLEAN DEFAULT 0);`
	CreateBusStopsSQL     = `CREATE TABLE IF NOT EXISTS bus_stops (bus_id INTEGER,station_id INTEGER, station_index INTEGER NOT NULL, PRIMARY KEY (bus_id, station_id), CONSTRAINT fk_bus_stops_busses FOREIGN KEY (bus_id) REFERENCES busses(id) ON DELETE CASCADE, CONSTRAINT fk_bus_stops_stations FOREIGN KEY (station_id) REFERENCES stations(id) ON DELETE CASCADE);`
	CreateStreetPointsSQL = `CREATE TABLE IF NOT EXISTS street_points (id INTEGER PRIMARY KEY,lat REAL,lng REAL,is_deleted BOOLEAN default false);`
	CreateStreetRelsSQL   = `CREATE TABLE IF NOT EXISTS street_rels (point_id INTEGER,point_index INTEGER,bus_id INTEGER,is_stop BOOLEAN default false, PRIMARY KEY (point_id, bus_id), CONSTRAINT fk_street_rels_busses FOREIGN KEY (bus_id) REFERENCES busses(id) ON DELETE CASCADE, CONSTRAINT fk_street_rels_street_points FOREIGN KEY (point_id) REFERENCES street_points(id) ON DELETE CASCADE);`
	CreateTimetablesSQL   = `CREATE TABLE IF NOT EXISTS time_tables (bus_id INTEGER, station_id INTEGER, enc_time INTEGER NOT NULL, CONSTRAINT fk_time_tables_busses FOREIGN KEY (bus_id) REFERENCES busses(id) ON DELETE CASCADE, CONSTRAINT fk_time_tables_stations FOREIGN KEY (station_id) REFERENCES stations(id) ON DELETE CASCADE);`

	BussesListSQL                        = `SELECT id, dir, name, from_station, to_station, no, color, website, urban, metropolitan, crawled FROM busses;`
	StationsListSQL                      = `SELECT id, name, street_name, lat, lng, outside, board FROM stations ORDER BY name ASC, street_name ASC;`
	NotCrawledBussesListSQL              = `SELECT id, dir, name, from_station, to_station, no, color, website, urban, metropolitan, crawled FROM busses WHERE crawled = false;`
	ShortBusByIDSQL                      = `SELECT b.id, b.dir, b.name, b.from_station, b.to_station, b.no, b.color, b.website, b.urban, b.metropolitan, b.crawled, s.id, s.name FROM busses b JOIN bus_stops bs ON b.id = bs.bus_id JOIN stations s ON bs.station_id = s.id WHERE b.id = ? ORDER BY bs.station_index;`
	FullBusByIDSQL                       = `SELECT b.id, b.dir, b.name, b.from_station, b.to_station, b.no, b.color, b.website, b.urban, b.metropolitan, b.crawled, s.id, s.name, s.lat, s.lng, s.outside, s.board, bs.station_index FROM busses b JOIN bus_stops bs ON b.id = bs.bus_id JOIN stations s ON bs.station_id = s.id WHERE b.id = ? ORDER BY bs.station_index;`
	GetTimeTablesForStationAndBusByIDSQL = `SELECT enc_time FROM time_tables WHERE station_id = ? AND bus_id = ?;`
	NotDeletedStreetPointsSQL            = `SELECT id, lat, lng FROM street_points WHERE is_deleted=false;`
	StreetRelsSQL                        = `SELECT point_id, bus_id, point_index, is_stop FROM street_rels ORDER BY point_index ASC;`
	CleanupStreetPointsSQL               = `DELETE FROM street_points WHERE 1 = 1;`
	InsertStreetRelsSQL                  = `INSERT INTO street_rels(point_id, point_index, bus_id, is_stop) VALUES (?, ?, ?, ?);`
	InsertStreetPointsSQL                = `INSERT INTO street_points(id, lat, lng) VALUES (?, ?, ?);`
	InsertTimeTablesSQL                  = `INSERT INTO time_tables (bus_id, station_id, enc_time) VALUES (?, ?, ?);`
	UpdateBussesSQL                      = `UPDATE busses SET crawled = true WHERE id = ?;`
	InsertBusSQL                         = `INSERT INTO busses (id, dir, name, from_station, to_station, no, color, website, urban, metropolitan, crawled) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	InsertBusStopSQL                     = `INSERT INTO bus_stops (bus_id, station_id, station_index) VALUES (?, ?, ?);`
	InsertStationSQL                     = `INSERT INTO stations (id, name, lat, lng, outside, board) VALUES (?, ?, ?, ?, ?, ?);`
	StationByIDSQL                       = `SELECT name, lat, lng, outside, board FROM stations WHERE id = ?;`
)

type LineNumberAndTime struct {
	No        string    `json:"b"`
	BusOSMID  int64     `json:"-"`
	Direction Direction `json:"-"`
	Times     []uint16  `json:"t,omitempty"`
}

type Lines []*LineNumberAndTime

type Station struct {
	OSMID         int64          `json:"i"`
	Index         int            `json:"d"`
	Name          string         `json:"n"`
	Street        sql.NullString `json:"-"`
	StreetName    string         `json:"s,omitempty"`
	Lat           float64        `json:"lt"`
	Lon           float64        `json:"ln"`
	HasBoard      bool           `json:"b,omitempty"`
	IsOutsideCity bool           `json:"o,omitempty"`
	IsTerminal    bool           `json:"t,omitempty"`
	Lines         Lines          `json:"l,omitempty"`
}

func (s Station) ID() int64 {
	return s.OSMID
}

func (l Lines) HasBus(busID int64) bool {
	for _, line := range l {
		if line.BusOSMID == busID {
			return true
		}
	}
	return false
}

func (l Lines) GetFirstEntry(busID int64) (*Time, bool) {
	for _, line := range l {
		if line.BusOSMID == busID {
			if len(line.Times) == 0 {
				return nil, false
			}
			result := Time{}
			result.Decompress(line.Times[0])
			return &result, true
		}
	}
	return nil, false
}

func (l Lines) GetFirstEntryAfter(busID int64, t *Time) (*Time, bool) {
	for _, line := range l {
		if line.BusOSMID == busID {
			if len(line.Times) == 0 {
				return nil, false
			}
			result := Time{}
			for _, ctime := range line.Times {
				result.Decompress(ctime)
				if t.After(result) {
					return &result, true
				}
			}
		}
	}
	return nil, false
}

type Busline struct {
	OSMID          int64     `json:"i"`
	Name           string    `json:"b"`
	From           string    `json:"f"`
	To             string    `json:"t"`
	Line           string    `json:"n"`
	Color          string    `json:"c"`
	Dir            int       `json:"d"`
	Link           string    `json:"w"`
	IsUrban        bool      `json:"u,omitempty"`
	IsMetropolitan bool      `json:"m,omitempty"`
	WasCrawled     bool      `json:"p"`
	Stations       []Station `json:"s"`
}

func (b *Busline) PrintStations() string {
	var sb strings.Builder
	for i, station := range b.Stations {
		if i > 0 {
			sb.WriteString("->")
		}
		sb.WriteString(fmt.Sprintf(" %d.[%d] %q", i+1, station.OSMID, station.Name))
	}
	return sb.String()
}

type Repository struct {
	*slog.Logger
	*sql.DB
}

func NewRepository(logger *slog.Logger, filePath string) (*Repository, error) {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return nil, err
	}

	for _, createTableSql := range []string{
		ForeignKeysSQL,
		CreateBussesSQL,
		CreateStationsSQL,
		CreateBusStopsSQL,
		CreateStreetPointsSQL,
		CreateStreetRelsSQL,
		CreateTimetablesSQL,
	} {
		_, err = db.Exec(createTableSql)
		if err != nil {
			logger.Error("error executing", "err", err, "sql", createTableSql)
			return nil, err
		}
	}

	return &Repository{Logger: logger, DB: db}, nil
}

func (r *Repository) GetBusses(notCrawled bool) ([]Busline, error) {
	SQL := BussesListSQL

	if notCrawled {
		SQL = NotCrawledBussesListSQL
	}

	rows, err := r.DB.Query(SQL)
	if err != nil {
		r.Logger.Error("error querying busses", "err", err)
		return nil, err
	}
	defer rows.Close()

	result := make([]Busline, 0)
	for rows.Next() {
		var b Busline
		err := rows.Scan(
			&b.OSMID,
			&b.Dir,
			&b.Name,
			&b.From,
			&b.To,
			&b.Line,
			&b.Color,
			&b.Link,
			&b.IsUrban,
			&b.IsMetropolitan,
			&b.WasCrawled,
		)
		if err != nil {
			r.Logger.Error("error scanning bus", "err", err)
			return nil, err
		}
		result = append(result, b)
	}

	if err := rows.Err(); err != nil {
		r.Logger.Error("error reading rows", "err", err)
		return nil, err
	}

	sort.Slice(result, func(i, j int) bool {
		return Compare(result[i].Line, result[j].Line) < 0
	})

	return result, nil
}

func (r *Repository) GetBusByID(id int64) (*Busline, error) {
	rows, err := r.DB.Query(ShortBusByIDSQL, id)
	if err != nil {
		r.Logger.Error("error querying bus", "id", id, "err", err)
		return nil, err
	}
	defer rows.Close()

	var result Busline
	for rows.Next() {
		var station Station
		err := rows.Scan(
			&result.OSMID,
			&result.Dir,
			&result.Name,
			&result.From,
			&result.To,
			&result.Line,
			&result.Color,
			&result.Link,
			&result.IsUrban,
			&result.IsMetropolitan,
			&result.WasCrawled,
			&station.OSMID,
			&station.Name,
		)
		if err != nil {
			r.Logger.Error("error scanning bus", "id", id, "err", err)
			return nil, err
		}
		result.Stations = append(result.Stations, station)
	}

	if err := rows.Err(); err != nil {
		r.Logger.Error("error reading rows", "id", id, "err", err)
		return nil, err
	}
	return &result, nil
}

func (r *Repository) GetFullStationsBusByID(id int64) (*Busline, error) {
	rows, err := r.DB.Query(FullBusByIDSQL, id)
	if err != nil {
		r.Logger.Error("error querying bus", "id", id, "err", err)
		return nil, err
	}
	defer rows.Close()

	var result Busline
	for rows.Next() {
		var station Station
		err := rows.Scan(
			&result.OSMID,
			&result.Dir,
			&result.Name,
			&result.From,
			&result.To,
			&result.Line,
			&result.Color,
			&result.Link,
			&result.IsUrban,
			&result.IsMetropolitan,
			&result.WasCrawled,
			&station.OSMID,
			&station.Name,
			&station.Lat,
			&station.Lon,
			&station.IsOutsideCity,
			&station.HasBoard,
			&station.Index,
		)
		if err != nil {
			r.Logger.Error("error scanning bus", "id", id, "err", err)

			return nil, err
		}
		result.Stations = append(result.Stations, station)
	}

	if err := rows.Err(); err != nil {
		r.Logger.Error("error reading rows", "id", id, "err", err)
		return nil, err
	}

	return &result, nil
}

func (r *Repository) GetTimeTablesForStationAndBus(stationID, busID int64) ([]uint16, error) {
	rows, err := r.DB.Query(GetTimeTablesForStationAndBusByIDSQL, stationID, busID)
	if err != nil {
		r.Logger.Error("error querying timetables", "err", err)
		return nil, err
	}
	defer rows.Close()

	result := make([]uint16, 0)
	for rows.Next() {
		var encodedTime uint16
		err := rows.Scan(&encodedTime)
		if err != nil {
			r.Logger.Error("error scanning encoded time", "err", err)
			return nil, err
		}
		result = append(result, encodedTime)
	}

	if err := rows.Err(); err != nil {
		r.Logger.Error("error reading rows", "err", err)
		return nil, err
	}

	return result, nil
}

func (r *Repository) GetStations() ([]Station, error) {
	rows, err := r.DB.Query(StationsListSQL)
	if err != nil {
		r.Logger.Error("error querying timetables", "err", err)
		return nil, err
	}
	defer rows.Close()

	result := make([]Station, 0)
	for rows.Next() {
		var station Station
		err := rows.Scan(
			&station.OSMID,
			&station.Name,
			&station.Street,
			&station.Lat,
			&station.Lon,
			&station.IsOutsideCity,
			&station.HasBoard,
		)
		if err != nil {
			r.Logger.Error("error scanning encoded time", "err", err)
			return nil, err
		}

		if station.Street.Valid {
			station.StreetName = station.Street.String
		}

		result = append(result, station)
	}

	if err := rows.Err(); err != nil {
		r.Logger.Error("error reading rows", "err", err)
		return nil, err
	}
	return result, nil
}

func (r *Repository) GetNotDeletedStreetPoints() (map[int64]Node, error) {
	rows, err := r.DB.Query(NotDeletedStreetPointsSQL)
	if err != nil {
		r.Logger.Error("error querying points", "err", err)
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]Node)
	for rows.Next() {
		var node Node
		err := rows.Scan(
			&node.ID,
			&node.Lat,
			&node.Lon,
		)
		if err != nil {
			r.Logger.Error("error scanning", "err", err)
			return nil, err
		}
		result[node.ID] = node
	}

	if err := rows.Err(); err != nil {
		r.Logger.Error("error reading rows", "err", err)
		return nil, err
	}

	return result, nil
}

func (r *Repository) GetStreetRels(includedPoints map[int64]Node) (map[int64][]Node, error) {
	if includedPoints == nil {
		return nil, errors.New("provide points")
	}

	rows, err := r.DB.Query(StreetRelsSQL)
	if err != nil {
		r.Logger.Error("error querying points", "err", err)
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64][]Node)
	for rows.Next() {
		var pointID, busID, pointIndex int64
		var isStop bool
		err := rows.Scan(
			&pointID,
			&busID,
			&pointIndex,
			&isStop,
		)
		if err != nil {
			r.Logger.Error("error scanning", "err", err)
			return nil, err
		}

		if _, has := result[busID]; !has {
			result[busID] = make([]Node, 0)
		}

		if point, has := includedPoints[pointID]; has {
			result[busID] = append(result[busID], Node{ID: pointID, Lat: point.Lat, Lon: point.Lon, Index: pointIndex, IsStop: isStop})
		}

	}

	if err := rows.Err(); err != nil {
		r.Logger.Error("error reading rows", "err", err)
		return nil, err
	}

	return result, nil
}

func (r *Repository) CleanupStreetPoints() error {
	_, err := r.DB.Exec(CleanupStreetPointsSQL)
	return err
}

func (r *Repository) LoadNewStreetPoints(bussesData Data) (map[int64][]Node, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return nil, err
	}

	pointsStmt, err := tx.Prepare(InsertStreetPointsSQL)
	if err != nil {
		return nil, err
	}
	defer pointsStmt.Close()

	uniqueWays := make(map[int64]Node)
	uniqueNodes := make(map[int64]Node)
	relationWays := make(map[int64][]Member)
	relationStops := make(map[int64][]Member)
	for _, element := range bussesData.Elements {
		if element.Type == OSMWay {
			uniqueWays[element.ID] = element
		} else if element.Type == OSMNode {
			uniqueNodes[element.ID] = element

			_, err = pointsStmt.Exec(element.ID, element.Lat, element.Lon)
			if err != nil {
				var sqliteErr sqlite3.Error
				if errors.As(err, &sqliteErr) {
					if !errors.Is(sqliteErr.Code, sqlite3.ErrConstraint) {
						if err := tx.Rollback(); err != nil {
							return nil, err
						}
						return nil, err
					}
				} else {
					r.Logger.Error("error inserting relation", "err", err)
					if err := tx.Rollback(); err != nil {
						return nil, err
					}
					return nil, err
				}
			}

		} else if element.Type == OSMRelation {
			// check excluded busses
			if _, willSkip := excludedBusses[element.ID]; willSkip {
				continue
			}

			refVal := element.Tags["ref"]
			if len(refVal) == 0 {
				continue
			}

			if strings.HasPrefix(refVal, "TE") || strings.HasPrefix(refVal, "XMAS") {
				continue
			}

			if refVal != "A1" {
				refVal = strings.ReplaceAll(refVal, "B", "")
				refVal = strings.ReplaceAll(refVal, "M", "")
				refVal = strings.ReplaceAll(refVal, "S", "")
				refVal = strings.ReplaceAll(refVal, "R", "")
				refVal = strings.ReplaceAll(refVal, "T", "")
				ref, refErr := strconv.ParseInt(refVal, 10, 64)
				if refErr != nil {
					r.Logger.Error("error parsing int", "err", err, "val", element.Tags["ref"])
					continue
				}

				// we skip points for metropolitan
				if ref > 100 {
					continue
				}
			}

			relationWays[element.ID] = make([]Member, 0)
			relationStops[element.ID] = make([]Member, 0)
			for _, member := range element.Members {
				if member.Type == OSMNode && member.Role == OSMStop {
					relationStops[element.ID] = append(relationStops[element.ID], member)
				}
				if member.Type == OSMWay {
					relationWays[element.ID] = append(relationWays[element.ID], member)
				}
			}
		}
	}

	relsStmt, err := tx.Prepare(InsertStreetRelsSQL)
	if err != nil {
		return nil, err
	}
	defer relsStmt.Close()

	processedWays := make(map[int64][]Node)
	for busID, ways := range relationWays {
		if len(ways) <= 0 {
			continue
		}

		processedWays[busID] = make([]Node, 0)
		pointIndex := 1
		lastNodeID := int64(-1)
		seen := make(map[int64]struct{})
		for _, wayMember := range ways {
			way, hasFoundWay := uniqueWays[wayMember.Ref]
			if !hasFoundWay {
				r.Logger.Error("ERROR FINDING WAY", "id", wayMember.Ref)
				continue
			}

			inReverse := false
			for wayNodeIndex, wayNodeID := range way.Nodes {
				if lastNodeID > 0 {
					if wayNodeIndex == 0 {
						if lastNodeID == wayNodeID {
							continue
						} else {
							inReverse = true
							break
						}
					}
				}

				lastNodeID = wayNodeID
				wayNode, wayNodeFound := uniqueNodes[wayNodeID]
				if !wayNodeFound {
					r.Logger.Error("WAY NODE NOT FOUND", "id", wayNodeID)
					break
				}

				itsAStop := false
				for _, stop := range relationStops[busID] {
					if stop.Ref == wayNodeID {
						itsAStop = true
						break
					}
				}

				if _, has := seen[wayNode.ID]; has {
					// r.Logger.Warn("has seen node", "id", wayNode.ID, "busID", busID, "stop", itsAStop)
					continue
				}

				_, err = relsStmt.Exec(wayNode.ID, pointIndex, busID, itsAStop)
				if err != nil {
					var sqliteErr sqlite3.Error
					if errors.As(err, &sqliteErr) {
						if !errors.Is(sqliteErr.Code, sqlite3.ErrConstraint) {
							if err := tx.Rollback(); err != nil {
								return nil, err
							}
							return nil, err
						}
					} else {
						r.Logger.Error("error inserting relation [1]", "err", err)
						if err := tx.Rollback(); err != nil {
							return nil, err
						}
						return nil, err
					}
				}

				seen[wayNode.ID] = struct{}{}
				pointIndex++
				// TODO : wayNode Type removal, is stop
				processedWays[busID] = append(processedWays[busID], wayNode)
			}

			if inReverse {
				for wayNodeIndex := len(way.Nodes) - 1; wayNodeIndex >= 0; wayNodeIndex-- {
					wayNodeID := way.Nodes[wayNodeIndex]
					lastNodeID = wayNodeID

					wayNode, wayNodeFound := uniqueNodes[wayNodeID]
					if !wayNodeFound {
						r.Logger.Error("[reverse] WAY NODE NOT FOUND", "id", wayNodeID)
						break
					}

					itsAStop := false
					for _, stop := range relationStops[busID] {
						if stop.Ref == wayNodeID {
							itsAStop = true
							break
						}
					}

					if _, has := seen[wayNode.ID]; has {
						// r.Logger.Warn("[reverse] has seen node", "id", wayNode.ID, "busID", busID, "stop", itsAStop)
						continue
					}

					_, err = relsStmt.Exec(wayNode.ID, pointIndex, busID, itsAStop)
					if err != nil {
						var sqliteErr sqlite3.Error
						if errors.As(err, &sqliteErr) {
							if !errors.Is(sqliteErr.Code, sqlite3.ErrConstraint) {
								if err := tx.Rollback(); err != nil {
									return nil, err
								}
								return nil, err
							}
						} else {
							r.Logger.Error("error inserting relation [2]", "err", err)
							if err := tx.Rollback(); err != nil {
								return nil, err
							}
							return nil, err
						}
					}

					seen[wayNode.ID] = struct{}{}
					pointIndex++
					// TODO : wayNode Type removal, is stop
					processedWays[busID] = append(processedWays[busID], wayNode)
				}
			}
		}
		r.Logger.Info("bus-to-points", "busID", busID, "points", pointIndex-1)
	}

	err = tx.Commit()
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return nil, err
		}
		return nil, err
	}

	return processedWays, nil
}

func (r *Repository) SaveTimeTables(busID int64, stationsTables map[int64]TimeTable) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	insertStmt, err := tx.Prepare(InsertTimeTablesSQL)
	if err != nil {
		return err
	}
	defer insertStmt.Close()

	for stationID, timeTable := range stationsTables {
		for _, t := range timeTable.RawTimes {
			_, err = insertStmt.Exec(busID, stationID, t.Compress())
			if err != nil {
				r.Logger.Error("error inserting into time tables", "err", err)
				if err := tx.Rollback(); err != nil {
					return err
				}
				continue
			}
		}

		r.Logger.Info("finished saving timetables between", "id", busID, "stationID", stationID, "len", len(timeTable.RawTimes))
	}

	err = tx.Commit()
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	_, err = r.DB.Exec(UpdateBussesSQL, busID)
	if err != nil {
		r.Logger.Error("error updating bus as crawled", "id", busID, "err", err)
		return err
	}

	return nil
}

func (r *Repository) CreateBus(bus *Busline) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	insertStmt, err := tx.Prepare(InsertBusSQL)
	if err != nil {
		return err
	}
	defer insertStmt.Close()

	_, err = insertStmt.Exec(bus.OSMID, bus.Dir, bus.Name, bus.From, bus.To, bus.Line, bus.Color, bus.Link, bus.IsUrban, bus.IsMetropolitan, false)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return nil
}

func (r *Repository) CreateStation(station *Station) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	insertStmt, err := tx.Prepare(InsertStationSQL)
	if err != nil {
		return err
	}
	defer insertStmt.Close()

	_, err = insertStmt.Exec(station.OSMID, station.Name, station.Lat, station.Lon, station.IsOutsideCity, station.HasBoard)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return nil
}

func (r *Repository) CreateStop(busId, stationId int64, stationIndex int) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	insertStmt, err := tx.Prepare(InsertBusStopSQL)
	if err != nil {
		return err
	}
	defer insertStmt.Close()

	_, err = insertStmt.Exec(busId, stationId, stationIndex)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	err = tx.Commit()
	if err != nil {
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	return nil
}

func (r *Repository) GetStationByID(stationId int64) (*Station, error) {
	var result Station
	err := r.DB.QueryRow(StationByIDSQL, stationId).Scan(&result.Name, &result.Lat, &result.Lon, &result.IsOutsideCity, &result.HasBoard)
	if err != nil {
		r.Logger.Error("error querying station", "err", err)
		return nil, err
	}
	return &result, nil
}

func (r *Repository) UpdateStreetName(stationId int64, name string) error {
	_, err := r.DB.Exec(`UPDATE stations SET street_name = ? WHERE id = ?;`, name, stationId)
	return err
}

func (r *Repository) GetStationsForBus(busId int64) ([]int64, error) {
	var result []int64

	rows, err := r.DB.Query(`SELECT station_id FROM bus_stops WHERE bus_id = ? ORDER BY station_index;`, busId)
	if err != nil {
		r.Logger.Error("error querying points", "err", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var stationID int64
		err := rows.Scan(&stationID)
		if err != nil {
			r.Logger.Error("error scanning", "err", err)
			return nil, err
		}
		result = append(result, stationID)
	}

	if err := rows.Err(); err != nil {
		r.Logger.Error("error reading rows", "err", err)
		return nil, err
	}

	return result, nil
}

func (r *Repository) GetAllTimetables() ([]Station, error) {
	rows, err := r.DB.Query(`SELECT bus_id, station_id, enc_time FROM time_tables ORDER BY station_id, bus_id, enc_time;`)
	if err != nil {
		r.Logger.Error("error querying points", "err", err)
		return nil, err
	}
	defer rows.Close()

	stationsMap := make(map[int64]*Station)

	for rows.Next() {
		var busID, stationID int64
		var encTime uint16
		err := rows.Scan(&busID, &stationID, &encTime)
		if err != nil {
			r.Logger.Error("error scanning", "err", err)
			return nil, err
		}

		station, has := stationsMap[stationID]
		if !has {
			station = &Station{OSMID: stationID}
		}

		busFound := false
		for _, line := range station.Lines {
			if line.BusOSMID == busID {
				busFound = true
				line.Times = append(line.Times, encTime)
				break
			}
		}

		if !busFound {
			line := LineNumberAndTime{BusOSMID: busID, Times: make([]uint16, 0)}
			line.Times = append(line.Times, encTime)
			station.Lines = append(station.Lines, &line)
		}

		if !has {
			stationsMap[stationID] = station
		}
	}

	if err := rows.Err(); err != nil {
		r.Logger.Error("error reading rows", "err", err)
		return nil, err
	}

	var result []Station
	for _, v := range stationsMap {
		result = append(result, *v)
	}

	return result, nil
}

func (r *Repository) GetAllFullBusses() ([]Busline, error) {
	const AllBussesSQL = `SELECT id, dir, name, from_station, to_station, no, color, website, urban, metropolitan, crawled FROM busses;`

	rows, err := r.DB.Query(AllBussesSQL)
	if err != nil {
		r.Logger.Error("error querying all busses", "err", err)
		return nil, err
	}
	defer rows.Close()

	busses := make(map[int64]*Busline)
	for rows.Next() {
		var bus Busline
		err := rows.Scan(
			&bus.OSMID,
			&bus.Dir,
			&bus.Name,
			&bus.From,
			&bus.To,
			&bus.Line,
			&bus.Color,
			&bus.Link,
			&bus.IsUrban,
			&bus.IsMetropolitan,
			&bus.WasCrawled,
		)
		if err != nil {
			r.Logger.Error("error scanning bus", "err", err)

			return nil, err
		}
		busses[bus.OSMID] = &bus
	}

	if err := rows.Err(); err != nil {
		r.Logger.Error("error reading rows", "err", err)
		return nil, err
	}

	const AllStopsSQL = `SELECT s.id, s.name, s.lat, s.lng, s.outside, s.board, s.street_name, bs.station_index, bs.bus_id FROM stations s JOIN bus_stops bs ON s.id = bs.station_id ORDER BY bs.station_index;`
	rows, err = r.DB.Query(AllStopsSQL)
	if err != nil {
		r.Logger.Error("error querying all full busses", "err", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var station Station
		var busID int64
		err := rows.Scan(
			&station.OSMID,
			&station.Name,
			&station.Lat,
			&station.Lon,
			&station.IsOutsideCity,
			&station.HasBoard,
			&station.Street,
			&station.Index,
			&busID,
		)
		if station.Street.Valid {
			station.StreetName = station.Street.String
		}

		if err != nil {
			r.Logger.Error("error scanning bus", "err", err)

			return nil, err
		}

		busses[busID].Stations = append(busses[busID].Stations, station)
	}

	var result []Busline
	for _, bus := range busses {
		result = append(result, *bus)
	}

	return result, nil
}
