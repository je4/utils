package ssh

import (
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"io"
)

type SFTPConnection struct {
	*Connection
	concurrency          int
	maxClientConcurrency int
	maxPacketSize        int
}

func NewSFTPConnection(sshConn *Connection, concurrency, maxClientConcurrency, maxPacketSize int) (*SFTPConnection, error) {
	sc := &SFTPConnection{
		Connection:           sshConn,
		concurrency:          concurrency,
		maxClientConcurrency: maxClientConcurrency,
		maxPacketSize:        maxPacketSize,
	}
	return sc, nil
}

func (sc *SFTPConnection) GetSFTPClient() (*sftp.Client, error) {
	sftpclient, err := sftp.NewClient(sc.Client, sftp.MaxPacket(sc.maxPacketSize), sftp.MaxConcurrentRequestsPerFile(sc.maxClientConcurrency))
	if err != nil {
		sc.Log.Infof("cannot get sftp subsystem - reconnecting to %s@%s", sc.Client.User(), sc.Address)
		if err := sc.Connect(); err != nil {
			return nil, errors.Wrapf(err, "cannot connect with ssh to %s@%s", sc.Client.User(), sc.Address)
		}
		sftpclient, err = sftp.NewClient(sc.Client)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create sftp client on %s@%s", sc.Client.User(), sc.Address)
		}
	}
	return sftpclient, nil
}

func (sc *SFTPConnection) ReadFile(path string, w io.Writer) (int64, error) {
	sftpclient, err := sc.GetSFTPClient()
	if err != nil {
		return 0, errors.Wrap(err, "unable to create SFTP session")
	}
	defer sftpclient.Close()

	r, err := sftpclient.Open(path)
	if err != nil {
		return 0, errors.Wrapf(err, "cannot open remote file %s", path)
	}
	defer r.Close()

	written, err := r.WriteTo(w) // io.Copy(w, r)
	if err != nil {
		return 0, errors.Wrap(err, "cannot sftpcopy data")
	}
	return written, nil
}

func (sc *SFTPConnection) WriteFile(path string, r io.Reader) (int64, error) {
	sftpclient, err := sc.GetSFTPClient()
	if err != nil {
		return 0, errors.Wrap(err, "unable to create SFTP session")
	}
	defer sftpclient.Close()

	w, err := sftpclient.Create(path)
	if err != nil {
		return 0, errors.Wrapf(err, "cannot create remote file %s", path)
	}

	written, err := w.ReadFromWithConcurrency(r, sc.concurrency) // io.Copy(w, r)
	if err != nil {
		return 0, errors.Wrap(err, "cannot sftpcopy data")
	}
	return written, nil
}
