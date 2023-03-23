package keepass2kms

import (
	"emperror.dev/errors"
	keepass "github.com/tobischo/gokeepasslib/v3"
	keepasswrappers "github.com/tobischo/gokeepasslib/v3/wrappers"
	"os"
	"strings"
)

func getRootGroup(grp *keepass.RootData, name string, create bool) *keepass.Group {
	for key, g := range grp.Groups {
		if g.Name == name {
			return &grp.Groups[key]
		}
	}
	if !create {
		return nil
	}
	grp.Groups = append(grp.Groups, keepass.Group{Name: name})
	return &grp.Groups[len(grp.Groups)-1]
}

func getSubGroup(grp *keepass.Group, name string, create bool) *keepass.Group {
	for key, g := range grp.Groups {
		if g.Name == name {
			return &grp.Groups[key]
		}
	}
	if !create {
		return nil
	}
	grp.Groups = append(grp.Groups, keepass.Group{Name: name})
	return &grp.Groups[len(grp.Groups)-1]
}

func getSubEntry(grp *keepass.Group, name string, create bool) *keepass.Entry {
	for key, e := range grp.Entries {
		if e.GetTitle() == name {
			if create {
				return nil
			} else {
				return &grp.Entries[key]
			}
		}
	}
	if !create {
		return nil
	}
	grp.Entries = append(grp.Entries, keepass.Entry{Values: []keepass.ValueData{mkValue("Title", name)}})
	return &grp.Entries[len(grp.Entries)-1]
}

func GetEntry(grp *keepass.RootData, name string, create bool) *keepass.Entry {
	parts := strings.Split(name, "/")
	group := getRootGroup(grp, parts[0], create)
	if group == nil {
		return nil
	}
	for i := 1; i < len(parts)-1; i++ {
		nextGroup := getSubGroup(group, parts[i], create)
		if nextGroup == nil {
			return nil
		}
		group = nextGroup
	}
	entry := getSubEntry(group, parts[len(parts)-1], create)
	if entry == nil {
		return nil
	}
	return entry
}

func mkValue(key string, value string) keepass.ValueData {
	return keepass.ValueData{Key: key, Value: keepass.V{Content: value}}
}

func mkProtectedValue(key string, value string) keepass.ValueData {
	return keepass.ValueData{
		Key:   key,
		Value: keepass.V{Content: value, Protected: keepasswrappers.NewBoolWrapper(true)},
	}
}

func LoadKeePassDBFromFile(file, credentials string) (*keepass.Database, error) {
	fp, err := os.Open(file)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open keePass file '%s'", file)
	}
	defer fp.Close()
	db := keepass.NewDatabase()
	db.Credentials = keepass.NewPasswordCredentials(credentials)
	if err := keepass.NewDecoder(fp).Decode(db); err != nil {
		return nil, errors.Wrapf(err, "cannot decode keePass file '%s'", file)
	}
	if err := db.UnlockProtectedEntries(); err != nil {
		return nil, errors.Wrap(err, "cannot unlock keepass db")
	}
	return db, nil
}
