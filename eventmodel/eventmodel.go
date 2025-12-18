package eventmodel

import (
	"fmt"
	"sort"
	"strings"
)

type EventModel struct {
	Slices     []Slice     `json:"slices"`
	Flows      []any       `json:"flows"`
	Aggregates []Aggregate `json:"aggregates"`
	Actors     []any       `json:"actors"`
	Context    string      `json:"context"`
	CodeGen    CodeGen     `json:"codeGen"`
	BoardID    string      `json:"boardId"`
}
type Field struct {
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	SubFields      []Field `json:"subfields"`
	Example        string  `json:"example"`
	Mapping        string  `json:"mapping"`
	Optional       bool    `json:"optional"`
	Generated      bool    `json:"generated"`
	Cardinality    string  `json:"cardinality"`
	IDAttribute    bool    `json:"idAttribute"`
	ExcludeFromAPI bool    `json:"excludeFromApi"`
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
type Readmodel struct {
	ID                    string         `json:"id"`
	Domain                string         `json:"domain"`
	ModelContext          string         `json:"modelContext"`
	Context               string         `json:"context"`
	Slice                 string         `json:"slice"`
	Title                 string         `json:"title"`
	Fields                []Field        `json:"fields"`
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
	Fields                []Field        `json:"fields"`
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
	Title    string  `json:"title"`
	ID       string  `json:"id"`
	Index    int     `json:"index"`
	Type     string  `json:"type"`
	Fields   []Field `json:"fields"`
	LinkedID string  `json:"linkedId"`
}
type Then struct {
	Title    string  `json:"title"`
	ID       string  `json:"id"`
	Index    int     `json:"index"`
	Type     string  `json:"type"`
	Fields   []Field `json:"fields"`
	LinkedID string  `json:"linkedId"`
}
type Comments struct {
	Description string `json:"description"`
}
type Specification struct {
	ID        string     `json:"id"`
	SliceName string     `json:"sliceName"`
	Title     string     `json:"title"`
	Given     []Given    `json:"given"`
	When      []Command  `json:"when"`
	Then      []Then     `json:"then"`
	Comments  []Comments `json:"comments"`
}
type Slice struct {
	ID             string          `json:"id"`
	Title          string          `json:"title"`
	Context        string          `json:"context"`
	Commands       []Command       `json:"commands"`
	Events         []Event         `json:"events"`
	Readmodels     []Readmodel     `json:"readmodels"`
	Screens        []Screens       `json:"screens"`
	Processors     []Processor     `json:"processors"`
	Specifications []Specification `json:"specifications"`
	Actors         []any           `json:"actors"`
	Aggregates     []string        `json:"aggregates"`
}

func (s *Slice) Instructions() string {
	var b strings.Builder

	for i, spec := range s.Specifications {
		b.WriteString(fmt.Sprintf("\n# Spec %d Start\n", i+1))
		b.WriteString(fmt.Sprintf("Title: %s", spec.Title))

		if len(spec.Comments) > 0 {
			b.WriteString("\nComments:\n")
			for _, c := range spec.Comments {
				b.WriteString(fmt.Sprintf("  - %s\n", c.Description))
			}
		}

		// Given
		if len(spec.Given) > 0 {
			b.WriteString("\n### Given (Events):\n")
			for _, e := range spec.Given {
				b.WriteString("  * " + e.Title)

				// Inline field analysis
				fieldsPrinted := false
				for _, field := range e.Fields {
					if field.Example != "" {
						if !fieldsPrinted {
							b.WriteString("\n  Fields:")
							fieldsPrinted = true
						}
						b.WriteString(fmt.Sprintf("\n   - %s: %s", field.Name, field.Example))
					}
				}
				b.WriteString("\n")
			}
		} else {
			b.WriteString("\n### Given (Events): None\n")
		}

		// When
		if len(spec.When) > 0 {
			b.WriteString("\n### When (Command):\n")
			for _, e := range spec.When {
				b.WriteString("  * " + e.Title)

				fieldsPrinted := false
				for _, field := range e.Fields {
					if field.Example != "" {
						if !fieldsPrinted {
							b.WriteString("\n  Fields:")
							fieldsPrinted = true
						}
						b.WriteString(fmt.Sprintf("\n   - %s: %s", field.Name, field.Example))
					}
				}
				b.WriteString("\n")
			}
		} else {
			b.WriteString("\n### When (Command): None\n")
		}

		// Then
		if len(spec.Then) > 0 {
			b.WriteString("\n### Then:\n")

			for _, e := range spec.Then {

				if e.Type == "SPEC_ERROR" {
					b.WriteString("\n### Then: Expect error\n")
					b.WriteString("  * " + e.Title)
				} else {
					b.WriteString("  * " + e.Title)

					fieldsPrinted := false
					for _, field := range e.Fields {
						if field.Example != "" {
							if !fieldsPrinted {
								b.WriteString("\n  Fields:")
								fieldsPrinted = true
							}
							b.WriteString(fmt.Sprintf("\n   - %s: %s", field.Name, field.Example))
						}
					}
				}
				b.WriteString("\n")
			}
		} else {
			b.WriteString("\n### Then: None\n")
		}

		b.WriteString(fmt.Sprintf("# Spec %d End\n", i+1))

	}

	return b.String()
}

type Aggregate struct {
	ID      string  `json:"id"`
	Title   string  `json:"title"`
	Fields  []Field `json:"fields"`
	Service any     `json:"service"`
	Type    string  `json:"type"`
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
	Fields                []Field        `json:"fields"`
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
	Fields                []Field        `json:"fields"`
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

func (c Command) ProducesEvents() []string {
	var procesEvent = make([]string, 0)
	for _, dependency := range c.Dependencies {
		if dependency.Type == "OUTBOUND" && dependency.ElementType == "EVENT" {
			procesEvent = append(procesEvent, dependency.ID)
		}
	}
	return procesEvent
}

type Processor struct {
	ID                    string         `json:"id"`
	Domain                string         `json:"domain"`
	ModelContext          string         `json:"modelContext"`
	Context               string         `json:"context"`
	Slice                 string         `json:"slice"`
	Title                 string         `json:"title"`
	Fields                []Field        `json:"fields"`
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

func (e EventModel) FindSlice(sliceName string) *Slice {
	for _, slice := range e.Slices {
		if slice.Title == sliceName {
			return &slice
		}
	}
	return nil
}

func (e EventModel) FindSliceByCommandId(id string) *Slice {
	for _, slice := range e.Slices {
		for _, cmd := range slice.Commands {
			if cmd.ID == id {
				return &slice
			}
		}
	}
	return nil
}

func (e EventModel) FindEventByID(id string) *Event {
	for _, slice := range e.Slices {
		for _, cmd := range slice.Events {
			if cmd.ID == id {
				return &cmd
			}
		}
	}
	return nil
}

func (e EventModel) FindSliceByReadModelId(id string) *Slice {
	for _, slice := range e.Slices {
		for _, rm := range slice.Readmodels {
			if rm.ID == id {
				return &slice
			}
		}
	}
	return nil
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

func SortFields(fields []Field) {
	sort.SliceStable(fields, func(i, j int) bool {
		a, b := fields[i], fields[j]

		// 1️⃣ ID attributes first
		if a.IDAttribute != b.IDAttribute {
			return a.IDAttribute
		}

		// 2️⃣ No cardinality before having one
		hasCardA := a.Cardinality != ""
		hasCardB := b.Cardinality != ""
		if hasCardA != hasCardB {
			return !hasCardA // no cardinality first
		}

		// 3️⃣ No subfields before having subfields
		hasSubA := len(a.SubFields) > 0
		hasSubB := len(b.SubFields) > 0
		if hasSubA != hasSubB {
			return !hasSubA
		}

		// 4️⃣ Alphabetical by name if otherwise equal
		return a.Name < b.Name
	})

	// Recurse into subfields too
	for i := range fields {
		if len(fields[i].SubFields) > 0 {
			SortFields(fields[i].SubFields)
		}
	}
}
