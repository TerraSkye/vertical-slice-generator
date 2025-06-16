package config

type Configuration struct {
	Slices     []Slices     `json:"slices"`
	Flows      []any        `json:"flows"`
	Aggregates []Aggregates `json:"aggregates"`
	Actors     []any        `json:"actors"`
	Context    string       `json:"context"`
	CodeGen    CodeGen      `json:"codeGen"`
	BoardID    string       `json:"boardId"`
}

func (c Configuration) GetTriggeringEvents() {

}

type Field struct {
	Name           string `json:"name"`
	Type           string `json:"type"`
	Example        string `json:"example"`
	Mapping        string `json:"mapping"`
	Optional       bool   `json:"optional"`
	Generated      bool   `json:"generated"`
	Cardinality    string `json:"cardinality"`
	IDAttribute    bool   `json:"idAttribute"`
	ExcludeFromAPI bool   `json:"excludeFromApi"`
}
type Dependencies struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Title       string `json:"title"`
	ElementType string `json:"elementType"`
}
type Prototype struct {
	ActiveByDefault bool `json:"activeByDefault"`
}
type Readmodels struct {
	ID                    string         `json:"id"`
	Domain                string         `json:"domain"`
	ModelContext          string         `json:"modelContext"`
	Context               string         `json:"context"`
	Slice                 string         `json:"slice"`
	Title                 string         `json:"title"`
	Fields                Fields         `json:"fields"`
	Type                  string         `json:"type"`
	Description           string         `json:"description"`
	Aggregate             string         `json:"aggregate"`
	AggregateDependencies []string       `json:"aggregateDependencies"`
	Dependencies          []Dependencies `json:"dependencies"`
	ListElement           bool           `json:"listElement"`
	APIEndpoint           string         `json:"apiEndpoint"`
	Service               any            `json:"service"`
	CreatesAggregate      bool           `json:"createsAggregate"`
	Prototype             Prototype      `json:"prototype"`
}
type Screens struct {
	ID                    string         `json:"id"`
	Domain                string         `json:"domain"`
	ModelContext          string         `json:"modelContext"`
	Context               string         `json:"context"`
	Slice                 string         `json:"slice"`
	Title                 string         `json:"title"`
	Fields                Fields         `json:"fields"`
	Type                  string         `json:"type"`
	Description           string         `json:"description"`
	Aggregate             string         `json:"aggregate"`
	AggregateDependencies []string       `json:"aggregateDependencies"`
	Dependencies          []Dependencies `json:"dependencies"`
	ListElement           bool           `json:"listElement"`
	APIEndpoint           string         `json:"apiEndpoint"`
	Service               any            `json:"service"`
	CreatesAggregate      bool           `json:"createsAggregate"`
	Prototype             Prototype      `json:"prototype"`
}
type Given struct {
	Title    string `json:"title"`
	ID       string `json:"id"`
	Index    int    `json:"index"`
	Type     string `json:"type"`
	Fields   Fields `json:"fields"`
	LinkedID string `json:"linkedId"`
}
type Then struct {
	Title    string `json:"title"`
	ID       string `json:"id"`
	Index    int    `json:"index"`
	Type     string `json:"type"`
	Fields   Fields `json:"fields"`
	LinkedID string `json:"linkedId"`
}
type Comments struct {
	Description string `json:"description"`
}
type Specifications struct {
	ID        string     `json:"id"`
	SliceName string     `json:"sliceName"`
	Title     string     `json:"title"`
	Given     []Given    `json:"given"`
	When      []Command  `json:"when"`
	Then      []Then     `json:"then"`
	Comments  []Comments `json:"comments"`
}
type Slices struct {
	ID             string           `json:"id"`
	Title          string           `json:"title"`
	Context        string           `json:"context"`
	Commands       []Command        `json:"commands"`
	Events         []Event          `json:"events"`
	Readmodels     []Readmodels     `json:"readmodels"`
	Screens        []Screens        `json:"screens"`
	Processors     []Processor      `json:"processors"`
	Specifications []Specifications `json:"specifications"`
	Actors         []any            `json:"actors"`
	Aggregates     []string         `json:"aggregates"`
}

type Aggregates struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Fields  Fields `json:"fields"`
	Service any    `json:"service"`
	Type    string `json:"type"`
}
type CodeGen struct {
	Application string `json:"application"`
	Domain      string `json:"domain"`
	RootPackage string `json:"rootPackage"`
}

type Event struct {
	ID                    string         `json:"id"`
	Domain                string         `json:"domain"`
	ModelContext          string         `json:"modelContext"`
	Context               string         `json:"context"`
	Slice                 string         `json:"slice"`
	Title                 string         `json:"title"`
	Fields                Fields         `json:"fields"`
	Type                  string         `json:"type"`
	Description           string         `json:"description"`
	Aggregate             string         `json:"aggregate"`
	AggregateDependencies []string       `json:"aggregateDependencies"`
	Dependencies          []Dependencies `json:"dependencies"`
	ListElement           bool           `json:"listElement"`
	//APIEndpoint           string         `json:"apiEndpoint"`
	Service          any       `json:"service"`
	CreatesAggregate bool      `json:"createsAggregate"`
	Prototype        Prototype `json:"prototype"`
}

type Command struct {
	ID                    string         `json:"id"`
	Domain                string         `json:"domain"`
	ModelContext          string         `json:"modelContext"`
	Context               string         `json:"context"`
	Slice                 string         `json:"slice"`
	Title                 string         `json:"title"`
	Fields                Fields         `json:"fields"`
	Type                  string         `json:"type"`
	Description           string         `json:"description"`
	Aggregate             string         `json:"aggregate"`
	AggregateDependencies []string       `json:"aggregateDependencies"`
	Dependencies          []Dependencies `json:"dependencies"`
	ListElement           bool           `json:"listElement"`
	APIEndpoint           string         `json:"apiEndpoint"`
	Service               any            `json:"service"`
	CreatesAggregate      bool           `json:"createsAggregate"`
	Prototype             Prototype      `json:"prototype"`
}

type Processor struct {
	ID                    string         `json:"id"`
	Domain                string         `json:"domain"`
	ModelContext          string         `json:"modelContext"`
	Context               string         `json:"context"`
	Slice                 string         `json:"slice"`
	Title                 string         `json:"title"`
	Fields                Fields         `json:"fields"`
	Type                  string         `json:"type"`
	Description           string         `json:"description"`
	Aggregate             string         `json:"aggregate"`
	AggregateDependencies []string       `json:"aggregateDependencies"`
	Dependencies          []Dependencies `json:"dependencies"`
	ListElement           bool           `json:"listElement"`
	APIEndpoint           string         `json:"apiEndpoint"`
	Service               any            `json:"service"`
	CreatesAggregate      bool           `json:"createsAggregate"`
	Prototype             Prototype      `json:"prototype"`
}

type Fields []Field

func (f Fields) IDAttributes() []Field {

	idAttributes := make([]Field, 0)

	for _, field := range f {

		if field.IDAttribute {
			idAttributes = append(idAttributes, field)
		}
	}

	if len(idAttributes) == 0 {
		for _, field := range f {
			if field.Name == "aggregateId" {
				idAttributes = append(idAttributes, field)
			}
		}
	}
	return idAttributes
}

func (f Fields) DataAttributes() []Field {

	idAttributes := make([]Field, 0)

	for _, field := range f {
		if !field.IDAttribute {
			idAttributes = append(idAttributes, field)
		}
	}

	if len(idAttributes) == len(f) {

		for idx := len(idAttributes) - 1; idx >= 0; idx-- {
			if idAttributes[idx].Name == "aggregateId" {
				idAttributes = append(idAttributes[:idx], idAttributes[idx+1:]...)
			}
		}
	}

	return idAttributes
}
