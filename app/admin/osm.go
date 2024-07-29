package admin

import (
	"io"
	"log"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/golang/geo/s2"
	"github.com/qedus/osmpbf"
)

const (
	tileSize = 2048

	DarkGrey     = "#1E232B"
	MediumGrey   = "#2A2E34"
	LightGrey    = "#3B3F46"
	DarkYellow   = "#EC9C04"
	MediumYellow = "#F5B301"
	LightYellow  = "#FED053"

	SpecialTag = "special"
	DefaultTag = "default"

	MaxLevel = 200
	MaxCells = 8000

	ItemTypeNode = iota
	ItemTypeWay
	ItemTypeRelation
)

type PBFData struct {
	Nodes  map[int64]Node
	Ways   map[int64]Way
	Findex S2Index
}

type FeatureRef struct {
	Id   int64
	Type int
	Zoom int
}

type S2Index map[s2.CellID][]FeatureRef

type IDraw interface {
	Draw(t *Tile)
}

func (o *PBFData) GetFeatures(northWestPoint, southEastPoint Node) []IDraw {
	var result []IDraw
	featureRefs := o.Findex.GetFeatures(northWestPoint, southEastPoint)

	// add non-special first
	for _, featureRef := range featureRefs {
		switch featureRef.Type {
		default:
		case ItemTypeWay:
			if _, isSpecial := o.Ways[featureRef.Id].Tags[SpecialTag]; !isSpecial {
				if _, isBoundary := o.Ways[featureRef.Id].Tags["boundary"]; !isBoundary {
					result = append(result, o.Ways[featureRef.Id])
				}
			}
		}
	}

	// add boundary next
	for _, featureRef := range featureRefs {
		switch featureRef.Type {
		default:
		case ItemTypeWay:
			if _, isSpecial := o.Ways[featureRef.Id].Tags[SpecialTag]; !isSpecial {
				if _, isBoundary := o.Ways[featureRef.Id].Tags["boundary"]; isBoundary {
					result = append(result, o.Ways[featureRef.Id])
				}
			}
		}
	}

	// add specials last
	for _, featureRef := range featureRefs {
		switch featureRef.Type {
		default:
		case ItemTypeWay:
			if _, isSpecial := o.Ways[featureRef.Id].Tags[SpecialTag]; isSpecial {
				result = append(result, o.Ways[featureRef.Id])
			}
		}
	}

	return result
}

func NodeFromPbf(n *osmpbf.Node) Node {
	return Node{Lon: n.Lon, Lat: n.Lat, ID: n.ID, Tags: n.Tags}
}

func WayFromPbf(w *osmpbf.Way) Way {
	return Way{NodeIDs: w.NodeIDs, Id: w.ID, Tags: w.Tags}
}

func RelationFromPbf(r *osmpbf.Relation) Node {
	result := Node{ID: r.ID, Tags: r.Tags}
	for _, member := range r.Members {
		memberType := "node"
		switch member.Type {
		case osmpbf.WayType:
			memberType = "way"
		case osmpbf.RelationType:
			memberType = "relation"
		}
		result.Members = append(result.Members, Member{Type: memberType, Ref: member.ID, Role: member.Role})
	}
	return result
}

func ReadPBFData(logger *slog.Logger, filename string, repo *Repository) (*PBFData, error) {
	mapRules := map[int][]Tag{
		0: {
			{"admin_level", "2"},
			{"natural", "coastline"},
		},
		3: {
			{"natural", "water"},
		},
		6: {
			{"admin_level", "3"},
			{"highway", "motorway"},
			{"highway", "trunk"},
			{"railway", "rail"},
		},
		8: {
			{"admin_level", "4"},
			{"highway", "primary"},
			{"highway", "secondary"},
			{"highway", "tertiary"},
		},
		10: {
			{"admin_level", "5"},
			{"highway", "motorway_link"},
			{"highway", "trunk_link"},
			{"highway", "primary_link"},
			{"highway", "secondary_link"},
			{"highway", "road"},
			{"railway", "light_rail"},
			{"railway", "monorail"},
		},
		12: {
			{"admin_level", "6"},
			{"admin_level", "7"},
			{"highway", "unclassified"},
			{"highway", "residential"},
			{"highway", "living_street"},
			{"highway", "bus_guideway"},
			{"highway", "raceway"},
		},
		13: {
			{"admin_level", "8"},
			{"highway", "bus_guideway"},
			{"highway", "highway_major_casing"},
		},
		14: {
			{"admin_level", "9"},
			{"highway", "bus_guideway"},
			{"highway", "primary"},
			{"highway", "highway_major_casing"},
		},
		15: {
			{"highway", "bus_guideway"},
			{"highway", "primary"},
			{"highway", "secondary"},
			{"highway", "highway_major_casing"},
		},
		16: {
			{"highway", "bus_guideway"},
			{"highway", "primary"},
			{"highway", "secondary"},
			{"highway", "tertiary"},
			{"highway", "highway_major_casing"},
		},
		17: {
			{"highway", "bus_guideway"},
			{"highway", "primary"},
			{"highway", "secondary"},
			{"highway", "tertiary"},
			{"highway", "service"},
			{"highway", "highway_major_casing"},
		},
	}

	data, err := ParsePbf(logger, filename, repo, mapRules)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ParsePbf(logger *slog.Logger, path string, repo *Repository, rules map[int][]Tag) (*PBFData, error) {
	start := time.Now()

	existingUrbanBusses := make(map[int64]Busline)
	busses, err := repo.GetBusses(false)
	if err != nil {
		panic(err)
	}

	for _, bus := range busses {
		if bus.IsUrban {
			existingUrbanBusses[bus.OSMID] = bus
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := osmpbf.NewDecoder(file)
	err = decoder.Start(runtime.GOMAXPROCS(0) - 1)
	if err != nil {
		return nil, err
	}

	data := &PBFData{Nodes: map[int64]Node{}, Ways: map[int64]Way{}, Findex: make(S2Index)}

	for {
		if v, err := decoder.Decode(); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		} else {
			switch v := v.(type) {
			case *osmpbf.Node:
				node := NodeFromPbf(v)
				data.Nodes[node.ID] = node
			case *osmpbf.Way:
				way := WayFromPbf(v)
				zoom, ok := way.MatchAny(rules)
				if ok {
					way.SetNodes(data.Nodes)
					data.Ways[way.Id] = way
					data.Findex.AddWay(way, zoom)
				}
			case *osmpbf.Relation:
				if v.Tags["network"] != "RAT Brașov" {
					continue
				}

				if _, has := existingUrbanBusses[v.ID]; !has {
					continue
				}

				for _, member := range v.Members {
					if member.Type != osmpbf.WayType {
						continue
					}

					if _, has := data.Ways[member.ID]; has {
						data.Ways[member.ID].Tags[SpecialTag] = SpecialTag
					} // ignore not found - metropolitan streets that were not imported (e.g. Predeal)
				}

			}
		}
	}

	log.Println("Num s2Cells", len(data.Findex))
	log.Println("Num ways", len(data.Ways))
	log.Println("Num nodes", len(data.Nodes))

	log.Println("took", time.Now().Sub(start).String())
	return data, nil
}

func (s S2Index) AddWay(way Way, zoom int) {
	capacity := s2.EmptyCap()

	for _, node := range way.Nodes {
		capacity = capacity.AddPoint(s2.PointFromLatLng(s2.LatLngFromDegrees(node.Lat, node.Lon)))
	}

	if capacity.IsEmpty() {
		return
	}

	region := &s2.RegionCoverer{MaxLevel: MaxLevel, MaxCells: MaxCells}
	cellUnion := region.FastCovering(capacity)
	for _, cellID := range cellUnion {
		s[cellID] = append(s[cellID], FeatureRef{way.Id, ItemTypeWay, zoom})
		for level := cellID.Level(); level > 0; level-- {
			cellID = cellID.Parent(level - 1)
			if _, ok := s[cellID]; !ok {
				s[cellID] = make([]FeatureRef, 0)
			}
		}
	}
}

func (s S2Index) GetFeatures(northWestPoint, southEastPoint Node) []FeatureRef {
	rectangle := s2.RectFromLatLng(s2.LatLngFromDegrees(northWestPoint.Lat, northWestPoint.Lon))
	rectangle = rectangle.AddPoint(s2.LatLngFromDegrees(southEastPoint.Lat, southEastPoint.Lon))

	region := &s2.RegionCoverer{MaxLevel: MaxLevel, MaxCells: MaxCells}

	cellUnion := region.Covering(rectangle)

	visitedCellIDs := make(map[s2.CellID]bool)
	nextVisit := make(map[int64]bool)
	result := make([]FeatureRef, 0)

	for _, currentID := range cellUnion {
		if featureRefs, ok := s[currentID]; ok {
			result = s.visitDown(currentID, featureRefs, nextVisit, result)
		}

		for level := currentID.Level(); level > 0; level-- {
			currentID = currentID.Parent(level - 1)
			result = s.visitUp(currentID, visitedCellIDs, nextVisit, result)
		}
	}

	return result
}

func (s S2Index) visitUp(cellID s2.CellID, visitedCellIDs map[s2.CellID]bool, nextVisit map[int64]bool, result []FeatureRef) []FeatureRef {
	featureRefs, ok := s[cellID]
	if !ok {
		return result
	}

	if visitedCellIDs[cellID] {
		return result
	}

	visitedCellIDs[cellID] = true
	for _, featureRef := range featureRefs {
		if !nextVisit[featureRef.Id] {
			result = append(result, featureRef)
			nextVisit[featureRef.Id] = true
		}
	}

	return result
}

func (s S2Index) visitDown(cellID s2.CellID, featureRefs []FeatureRef, nextVisit map[int64]bool, result []FeatureRef) []FeatureRef {
	for _, featureRef := range featureRefs {
		if !nextVisit[featureRef.Id] {
			result = append(result, featureRef)
			nextVisit[featureRef.Id] = true
		}
	}

	if !cellID.IsLeaf() {
		for _, childCellID := range cellID.Children() {
			if v, ok := s[childCellID]; ok {
				result = s.visitDown(childCellID, v, nextVisit, result)
			}
		}
	}

	return result
}
