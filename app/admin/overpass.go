package admin

import (
	"strings"
)

const (
	OSMNode     = "node"
	OSMStop     = "stop"
	OSMWay      = "way"
	OSMRelation = "relation"
)

type Direction int

const (
	NotSet Direction = 0
	Tour   Direction = 1
	Retour Direction = 2
)

func (d Direction) String() string {
	switch d {
	case Tour:
		return "dus"
	case Retour:
		return "intors"
	}
	return "NOT SET"
}

type Data struct {
	Elements []Node `json:"elements"`
}

type Member struct {
	Type string `json:"type"`
	Ref  int64  `json:"ref"`
	Role string `json:"role,omitempty"`
}

type Node struct {
	Type    string            `json:"type,omitempty"`    //
	ID      int64             `json:"id,omitempty"`      //
	Index   int64             `json:"idx,omitempty"`     //
	Lat     float64           `json:"lat"`               //
	Lon     float64           `json:"lon"`               //
	Tags    map[string]string `json:"tags,omitempty"`    //
	Members []Member          `json:"members,omitempty"` // relation members
	Nodes   []int64           `json:"nodes,omitempty"`   // way nodes
	IsStop  bool              `json:"stop,omitempty"`    //
}

func (n *Node) CleanupTags() {
	for k, v := range n.Tags {
		switch k {
		case "bench", "bus", "bin", "shelter", "wheelchair", "network", "network:wikidata", "network:wikipedia", "operator":
			delete(n.Tags, k)
		case "public_transport":
			if v == "platform" {
				delete(n.Tags, k)
			}
		case "highway":
			if v == "bus_stop" {
				delete(n.Tags, k)
			}
		}
	}
}

func NewStation(fromNode Node) Station {
	result := Station{
		OSMID: fromNode.ID,
		Lat:   fromNode.Lat,
		Lon:   fromNode.Lon,
	}
	for k, v := range fromNode.Tags {
		switch k {
		case "departures_board":
			if v == "realtime" {
				result.HasBoard = true
			}
		case "fare_zone":
			if len(v) > 0 {
				result.IsOutsideCity = true
			}
		case "name":
			result.Name = v
		}
	}
	return result
}

func NewBusline(fromNode Node) Busline {
	result := Busline{OSMID: fromNode.ID}

	overriddenLine := ""
	for k, v := range fromNode.Tags {
		switch k {
		case "name":
			cleanName := strings.ReplaceAll(v, "=>", "-")
			result.Name = cleanName
		case "from":
			result.From = v
		case "to":
			result.To = v
		case "colour":
			result.Color = v
		case "ref":
			result.Line = v
		case "local_ref":
			overriddenLine = v
		case "website":
			result.Link = v
		}
	}

	if len(overriddenLine) > 0 {
		result.Line = overriddenLine
	}

	if strings.Contains(result.Link, "-dus") {
		result.Dir = 1
	} else if strings.Contains(result.Link, "-intors") {
		result.Dir = 2
	}

	return result
}
