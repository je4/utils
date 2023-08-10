package marc21schema

import (
	_ "embed"
	"encoding/json"
	"log"
	"sync"
)

//go:embed schema.json
var schemaJSON []byte

type Label struct {
	Label string `json:"label"`
}

type Position struct {
	Label             string            `json:"label"`
	Url               string            `json:"url,omitempty"`
	Start             int               `json:"start"`
	End               int               `json:"end"`
	RepeatableCOntent bool              `json:"repeatableCOntent"`
	Solr              string            `json:"solr,omitempty"`
	Codes             map[string]*Label `json:"codes,omitempty"`
	HistoricalCodes   map[string]*Label `json:"historical-codes,omitempty"`
}

type Positions map[string]*Position

func (ps Positions) ToSolr(str string) map[string][]any {
	var l = len(str)
	var result = map[string][]any{}
	for _, pos := range ps {
		if l < pos.End-1 {
			break
		}
		start := pos.Start
		end := pos.End
		if start >= len(str) {
			continue
		}
		if end > len(str) {
			continue
		}

		code := str[start:end]
		result[pos.Solr] = []any{code}
		/*
			label, ok := pos.Codes[code]
			if !ok {
				label, ok = pos.HistoricalCodes[code]
				if !ok {
					label = &Label{Label: code}
				}
			}
			result[pos.Solr] = label.Label

		*/
	}
	return result
}

type Type struct {
	Positions Positions `json:"positions,omitempty"`
}

type CodeList struct {
	Name  string            `json:"name"`
	Url   string            `json:"url"`
	Codes map[string]*Label `json:"codes"`
}

type Subfield struct {
	Label      string    `json:"label,omitempty"`
	Repeatable bool      `json:"repeatable"`
	Solr       string    `json:"solr,omitempty"`
	CodeList   *CodeList `json:"codelist,omitempty"`
}

type Indicator struct {
	Label string            `json:"label"`
	Codes map[string]*Label `json:"codes"`
}

type Field struct {
	Repeatable bool                 `json:"repeatable"`
	Positions  Positions            `json:"positions,omitempty"`
	Tag        string               `json:"tag,omitempty"`
	Label      string               `json:"label,omitempty"`
	Solr       string               `json:"solr"`
	Types      map[string]*Type     `json:"types,omitempty"`
	Ind1       *Indicator           `json:"indicator1,omitempty"`
	Ind2       *Indicator           `json:"indicator2,omitempty"`
	Subfields  map[string]*Subfield `json:"subfields,omitempty"`
}

/*
func (flds Field) toSolr(df *marc21.Datafield) (result any, err error) {
	if len(df.Subfields) > 0 {

	}
	return
}

*/

type FieldList map[string]*Field

/*
func (flds FieldList) toSolr(dfs []*marc21.Datafield) (result map[string][]any, err error) {
	for _, df := range dfs {
		fldSchema, ok := flds[df.GetTag()]
		if !ok {
			return nil, fmt.Errorf("no schema for field %s", df.GetTag())
		}
		if fldSchema.Tag != df.GetTag() {
			return nil, fmt.Errorf("invalid field - %s != %s", fldSchema.Tag, df.GetTag())
		}
		if _, ok := result[fldSchema.Solr]; !ok {
			result[fldSchema.Solr] = []any{}
		}
		r, err := fldSchema.toSolr(df)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create solr struct for %s", df.GetTag())
		}
		result[fldSchema.Solr] = append(result[fldSchema.Solr], r)
	}
	return result, nil
}

*/

type Schema struct {
	Schema      string    `json:"$schema"`               // "$schema": "https://format.gbv.de/schema/avram/schema.json",
	Title       string    `json:"title"`                 // "title": "MARC 21 Format for Bibliographic Data.",
	Description string    `json:"description,omitempty"` // "description": "MARC 21 Format for Bibliographic Data.",
	URL         string    `json:"url,omitempty"`         //"url": "https://www.loc.gov/marc/bibliographic/",
	Fields      FieldList `json:"fields"`                // "fields":
}

/*
func (schema *Schema) ToSolr(rec marc21.Record) (map[string]any, error) {

	return nil, nil
}

*/

var schema *Schema

var sLock sync.Mutex

func GetSchema() *Schema {
	sLock.Lock()
	defer sLock.Unlock()
	if schema == nil {
		schema = &Schema{}
		if err := json.Unmarshal(schemaJSON, schema); err != nil {
			log.Fatal(err)
		}
	}
	return schema
}

/*
func GetLDR(record marc21.Record) map[string][]any {
	str := record.Leader.GetText()
	schema := GetSchema()
	ldrSchema := schema.Fields["LDR"]
	return ldrSchema.Positions.ToSolr(str)
}

*/
