package maps

type Location struct {
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
}

type DirectionsResult struct {
	From        Location    `json:"from"`
	To          Location    `json:"to"`
	Mode        string      `json:"mode"`
	Distance    string      `json:"distance"`
	Duration    string      `json:"duration"`
	Steps       []RouteStep `json:"steps"`
	RouteCoords [][]float64 `json:"route_coords"`
}

type RouteStep struct {
	Instruction string `json:"instruction"`
	Distance    string `json:"distance"`
}

type GeocodeResult struct {
	Location Location `json:"location"`
}

type OSRMResponse struct {
	Code   string      `json:"code"`
	Routes []OSRMRoute `json:"routes"`
}

type OSRMRoute struct {
	Distance float64   `json:"distance"`
	Duration float64   `json:"duration"`
	Geometry string    `json:"geometry"`
	Legs     []OSRMLeg `json:"legs"`
}

type OSRMLeg struct {
	Steps []OSRMStep `json:"steps"`
}

type OSRMStep struct {
	Distance float64      `json:"distance"`
	Duration float64      `json:"duration"`
	Name     string       `json:"name"`
	Maneuver OSRMManeuver `json:"maneuver"`
}

type OSRMManeuver struct {
	Type     string    `json:"type"`
	Modifier string    `json:"modifier"`
	Location []float64 `json:"location"`
}
