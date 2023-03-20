package keepass2kms

import (
	"encoding/base64"
	"github.com/pkg/errors"
	"github.com/tobischo/gokeepasslib/v3"
	"os"
)

type Keepass2 struct {
	filename        string
	db              *gokeepasslib.Database
	backupExtension string
}

func NewKeepass2(filename string, credentials string, backupExtension string) (*Keepass2, error) {
	kp2 := &Keepass2{
		filename:        filename,
		backupExtension: backupExtension,
	}
	if err := kp2.Open(credentials); err != nil {
		return nil, errors.Wrapf(err, "cannot open keepass2 database %s", filename)
	}
	return kp2, nil
}

func (k *Keepass2) Open(credentials string) error {
	fp, err := os.Open(k.filename)
	if err != nil {
		return errors.Wrapf(err, "cannot open keepass2 database %s", k.filename)
	}
	defer fp.Close()
	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials(credentials)
	if err := gokeepasslib.NewDecoder(fp).Decode(db); err != nil {
		return errors.Wrapf(err, "cannot decode keepass2 database %s", k.filename)
	}
	if err := db.UnlockProtectedEntries(); err != nil {
		return errors.Wrapf(err, "cannot unlock keepass2 database %s", k.filename)
	}
	k.db = db
	return nil
}

func (k *Keepass2) Close() error {
	if err := k.db.LockProtectedEntries(); err != nil {
		return errors.Wrapf(err, "cannot lock keepass2 database %s", k.filename)
	}
	if k.backupExtension != "" {
		if err := os.Rename(k.filename, k.filename+k.backupExtension); err != nil {
			return errors.Wrapf(err, "cannot rename keepass2 database '%s' -> '%s'", k.filename, k.filename+k.backupExtension)
		}
	}
	fp, err := os.Create(k.filename)
	if err != nil {
		return errors.Wrapf(err, "cannot create keepass2 database '%s'", k.filename)
	}
	defer fp.Close()
	if err := gokeepasslib.NewEncoder(fp).Encode(k.db); err != nil {
		return errors.Wrapf(err, "cannot encode keepass2 database %s", k.filename)
	}
	return nil
}

func (k *Keepass2) GetEntry(key string) (*gokeepasslib.Entry, error) {
	entry := getEntry(k.db.Content.Root, key, false)
	if entry == nil {
		return nil, errors.Errorf("cannot get entry '%s'", key)
	}
	return entry, nil
}

func (k *Keepass2) NewEntry(key string, data []byte, associatedData []byte, nonce []byte) error {
	entry := getEntry(k.db.Content.Root, key, true)
	if entry == nil {
		return errors.Errorf("entry '%s' already exists", key)
	}
	entry.Values = append(entry.Values, mkProtectedValue("Password", base64.StdEncoding.EncodeToString(associatedData)))
	entry.Values = append(entry.Values, mkProtectedValue("Key", base64.StdEncoding.EncodeToString(data)))
	entry.Values = append(entry.Values, mkProtectedValue("Nonce", base64.StdEncoding.EncodeToString(nonce)))
	return nil
}
