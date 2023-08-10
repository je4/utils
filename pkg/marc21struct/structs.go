package marc21struct

import (
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"gitlab.switch.ch/ub-unibas/alma2elastic/v2/pkg/marc21schema"
	"golang.org/x/text/unicode/norm"
	"regexp"
	"strings"
	"time"
)

/*
   MARC-8 vs. UTF-8 encoding
       * leader.CharacterCodingScheme == "a" is UCS/Unicode
       * https://www.loc.gov/marc/specifications/speccharucs.html
       * https://www.loc.gov/marc/specifications/codetables.xml
       * https://lcweb2.loc.gov/diglib/codetables/eacc2uni.txt
*/

/*
https://www.loc.gov/marc/specifications/specrecstruc.html
*/

const (
	delimiter        = byte(0x1f)
	fieldTerminator  = byte(0x1e)
	recordTerminator = byte(0x1d)
	leaderLen        = 24
	maxRecordSize    = 99999
)

// Leader is for containing the text string of the MARC record Leader
type Leader struct {
	Text string `xml:",chardata" json:"text"`
}

func (ldr Leader) ToSolr() map[string][]any {
	schema := marc21schema.GetSchema()
	ldrField := schema.Fields["LDR"]
	return ldrField.Positions.ToSolr(ldr.Text)
}

type Date time.Time

func (d *Date) MarshalJSON() ([]byte, error) {
	str := (*time.Time)(d).Format("2006-01-02")
	return json.Marshal(str)
}

type DateTime time.Time

func (d *DateTime) MarshalJSON() ([]byte, error) {
	str := (*time.Time)(d).Format("2006-01-02 15:04:05")
	return json.Marshal(str)
}

type Controlfields []*Controlfield

func (cfs Controlfields) MarshalJSON() ([]byte, error) {
	result := map[string]string{}
	for _, cf := range cfs {
		l := len(cf.Text)
		switch cf.Tag {
		case "001":
			result["001_id"] = cf.Text
		case "003":
			result["003_controlNumberId"] = cf.Text
		case "005":
			result["005_latestTransactionTime"] = cf.Text
		case "006":
			result["006_full"] = cf.Text
		case "007":
			result["007_full"] = cf.Text
		case "008":
			if l != 40 {
				matched, _ := regexp.MatchString(`^[0-9]{6}[a-z][0-9u |]{8}[a-z].{19}[a-z|]{3}`, cf.Text)
				if matched != true {
					return nil, errors.Errorf("controlfield invalid - '%s'", cf.Text)
				}
			}
			result["008_full"] = cf.Text
			result["008_06_typeOfDate"] = cf.Text[6:7]
			result["008_07-10_dateFirst"] = cf.Text[7:11]
			result["008_11-14_dateSecond"] = cf.Text[11:15]
			result["008_15-17_country"] = cf.Text[15:18]
			result["008_35-37_language"] = cf.Text[35:38]
			switch cf.Type {
			case "Books", "Computer Files", "Music", "Continuing Resources", "Mixed Materials":
				result["008_23or29_formOfItem"] = cf.Text[23:24]
			case "Visual Materials", "Maps":
				result["008_23or29_formOfItem"] = cf.Text[29:30]
			default:
			}
		default:
		}
	}
	return json.Marshal(result)
}

// Record is for containing a MARC record
type Record struct {
	Leader        Leader        `xml:"leader" json:"LDR"`
	Controlfields Controlfields `xml:"controlfield" json:"controlfield"`
	Datafields    []*Datafield  `xml:"datafield" json:"datafield"`
}

// Controlfield contains a controlfield entry
type Controlfield struct {
	Tag  string `xml:"tag,attr" json:"tag"`
	Text string `xml:",chardata" json:"text"`
	Type string `xml:"-" json:"-"`
}

func (ctrl Controlfield) buildFields() map[string][]any {
	var fields map[string][]any
	schema := marc21schema.GetSchema()
	ctrlField := schema.Fields[ctrl.Tag]

	var fldname string
	if ctrlField != nil && ctrlField.Solr != "" {
		fldname = ctrlField.Solr
	} else {
		fldname = ctrl.Tag
	}
	fields = map[string][]any{fldname: []any{ctrl.Text}}

	if ctrlField != nil {
		if ctrlField.Positions == nil {
			if ctrlField.Types != nil && ctrl.Type != "" {
				if t, ok := ctrlField.Types["All Materials"]; ok {
					if p := t.Positions; p != nil {
						fields = p.ToSolr(ctrl.Text)
					}
				}
				if t, ok := ctrlField.Types[ctrl.Type]; ok {
					if p := t.Positions; p != nil {
						for k, v := range p.ToSolr(ctrl.Text) {
							fields[k] = v
						}
					}
				}
			} else {
			}
		} else {
			fields = ctrlField.Positions.ToSolr(ctrl.Text)
		}
	}
	return fields
}

func (ctrl Controlfield) MarshalJSON() ([]byte, error) {
	fields := ctrl.buildFields()
	return json.Marshal(fields)
}

// Implement the Stringer interface for "Pretty-printing"
func (cf Controlfield) String() string {
	return fmt.Sprintf("{%s: '%s'}", cf.Tag, cf.Text)
}

// Datafield contains a datafield entry
type Datafield struct {
	Tag       string      `xml:"tag,attr" json:"tag"`
	Ind1      string      `xml:"ind1,attr" json:"ind1,omitempty"`
	Ind2      string      `xml:"ind2,attr" json:"ind2,omitempty"`
	Subfields []*Subfield `xml:"subfield" json:"subfield,omitempty"`
}

func (df Datafield) ToSolr() (result map[string][]any) {
	result = map[string][]any{}
	schema := marc21schema.GetSchema()
	fldSchema := schema.Fields[df.Tag]
	if df.Subfields != nil {
		for _, sf := range df.Subfields {
			var fldname = strings.ReplaceAll("unknown_"+df.Tag+df.Ind1+df.Ind2+sf.Code, " ", "_")
			if fldSchema != nil {
				sfSchema := fldSchema.Subfields[sf.Code]
				if sfSchema != nil {
					fldname = strings.ReplaceAll(sfSchema.Solr[0:3]+df.Ind1+df.Ind2+sfSchema.Solr[3:], " ", "_")
				}
			}
			if _, ok := result[fldname]; !ok {
				result[fldname] = []any{}
			}
			result[fldname] = append(result[fldname], sf.Text)
		}
	}
	return
}

// GetSubfields returns subfields for the datafield that match the
// specified codes. If no codes are specified (empty string) then all
// subfields are returned
func (df Datafield) GetSubfields(codes string) (sfs []*Subfield) {
	if codes == "" {
		return df.Subfields
	}

	for _, c := range []byte(codes) {
		for _, sf := range df.Subfields {
			if sf.Code == string(c) {
				sfs = append(sfs, sf)
			}
		}
	}
	return sfs
}

// Subfield contains a subfield entry
type Subfield struct {
	Code      string `xml:"code,attr" json:"code"`
	Text      string `xml:",chardata" json:"text"`
	Datafield *Datafield
}

func (sub *Subfield) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("|%s %s", sub.Code, sub.Text))
	//var subMap = map[string]string{sub.Code: sub.Text}
	//return json.Marshal(subMap)
}

var subfieldRegexp = regexp.MustCompile(`^\\|(.) (.+)$`)

func (sub *Subfield) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return errors.WithStack(err)
	}
	matches := subfieldRegexp.FindStringSubmatch(str)
	if matches == nil {
		return errors.Errorf("invalid subfield content '%s'", string(data))
	}
	sub.Code = matches[1]
	sub.Text = matches[2]
	/*
		var subMap = map[string]string{}
		if err := json.Unmarshal(data, &subMap); err != nil {
			return errors.WithStack(err)
		}
		for key, val := range subMap {
			sub.Code = key
			sub.Text = val
			break
		}
	*/
	return nil
}

func (sub *Subfield) GetMarc() string {
	sb := strings.Builder{}
	sb.WriteString(sub.Datafield.Tag)
	sb.WriteString(sub.Datafield.Ind1)
	sb.WriteString(sub.Datafield.Ind2)
	sb.WriteString(sub.Code)
	return sb.String()
}

func (sub *Subfield) Decompose() {
	sub.Text = norm.NFC.String(sub.Text)
}

type QueryStruct struct {
	Name          string          `json:"name"`
	Field         *Datafield      `json:"field"`
	Datafields    []*Datafield    `json:"datafield"`
	Controlfields []*Controlfield `json:"controlfield"`
	Leader        string          `json:"LDR"`
	//	tagRef        map[string]int  `json:"-"`
}

type QueryStructMARCIJ struct {
	Name   string       `json:"name"`
	Field  *MARCIJField `json:"field"`
	Object *MARCIJ      `json:"object"`
}
