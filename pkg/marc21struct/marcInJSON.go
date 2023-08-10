package marc21struct

import (
	"emperror.dev/errors"
	"encoding/json"
)

type MARCIJContent struct {
	Ind1      string              `json:"ind1,omitempty"`
	Ind2      string              `json:"ind2,omitempty"`
	Subfields []map[string]string `json:"subfields"`
}

type MARCIJField struct {
	MARCIJContent
	Code string
	Text string
}

func (f *MARCIJField) MarshalJSON() ([]byte, error) {
	if f.Text != "" {
		fld := map[string]string{f.Code: f.Text}
		return json.Marshal(fld)
	} else {
		fld := map[string]MARCIJContent{f.Code: f.MARCIJContent}
		return json.Marshal(fld)
	}
}

func (f *MARCIJField) UnmarshalJSON(data []byte) error {
	var fld = map[string]string{}
	if err := json.Unmarshal(data, &fld); err != nil {
		var fld = map[string]MARCIJContent{}
		if err2 := json.Unmarshal(data, &fld); err2 != nil {
			return errors.WithStack(errors.Combine(err, err2))
		}
		for code, content := range fld {
			f.Code = code
			f.MARCIJContent = content
			break
		}
		return nil
	} else {
		for code, text := range fld {
			f.Code = code
			f.Text = text
			break
		}
		return nil
	}
}

func (f *MARCIJField) fromMarcControlfield(controlField *Controlfield) error {
	f.MARCIJContent = MARCIJContent{}
	f.Code = controlField.Tag
	f.Text = controlField.Text
	return nil
}

func (f *MARCIJField) FromMarc(dataField *Datafield) error {
	f.MARCIJContent = MARCIJContent{}
	f.Code = dataField.Tag
	f.MARCIJContent = MARCIJContent{
		Ind1:      dataField.Ind1,
		Ind2:      dataField.Ind2,
		Subfields: []map[string]string{},
	}
	for _, field := range dataField.Subfields {
		f.Subfields = append(f.Subfields, map[string]string{field.Code: field.Text})
	}
	return nil
}

type MARCIJ struct {
	Leader string         `json:"leader"`
	Fields []*MARCIJField `json:"fields"`
}

func (pmr *MARCIJ) FromMarc(mr *Record) error {
	pmr.Leader = mr.Leader.Text
	for _, controlField := range mr.Controlfields {
		fld := &MARCIJField{}
		if err := fld.fromMarcControlfield(controlField); err != nil {
			return errors.WithStack(err)
		}
		pmr.Fields = append(pmr.Fields, fld)
	}
	for _, dataField := range mr.Datafields {
		pmf := &MARCIJField{}
		if err := pmf.FromMarc(dataField); err != nil {
			return errors.WithStack(err)
		}
		pmr.Fields = append(pmr.Fields, pmf)
	}
	return nil
}
