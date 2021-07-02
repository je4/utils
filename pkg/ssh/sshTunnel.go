package ssh

/*
https://gist.github.com/svett/5d695dcc4cc6ad5dd275

*/

import (
	"emperror.dev/errors"
	"fmt"
	"github.com/op/go-logging"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"sync"
)

type Endpoint struct {
	Host string
	Port int
}

type SourceDestination struct {
	Local  *Endpoint
	Remote *Endpoint
}

func (endpoint *Endpoint) String() string {
	return fmt.Sprintf("%s:%d", endpoint.Host, endpoint.Port)
}

type SSHtunnel struct {
	server   *Endpoint
	pool     *ConnectionPool
	tunnels  map[string]*SourceDestination
	quit     map[string]chan interface{}
	listener map[string]net.Listener
	config   *ssh.ClientConfig
	client   *ssh.Client
	log      *logging.Logger
	wg       sync.WaitGroup
}

func NewSSHTunnel(user, privateKey string, serverEndpoint *Endpoint, tunnels map[string]*SourceDestination, log *logging.Logger) (*SSHtunnel, error) {
	key, err := ioutil.ReadFile(privateKey)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to read private key %s", privateKey)
	}
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to parse private key")
	}

	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	tunnel := &SSHtunnel{
		config:   sshConfig,
		server:   serverEndpoint,
		tunnels:  tunnels,
		pool:     NewConnectionPool(log),
		quit:     make(map[string]chan interface{}),
		listener: make(map[string]net.Listener),
		log:      log,
	}

	return tunnel, nil
}

func (tunnel *SSHtunnel) String() string {
	str := fmt.Sprintf("%v@%v:%v",
		tunnel.config.User,
		tunnel.server.Host, tunnel.server.Port,
	)
	for _, srcdests := range tunnel.tunnels {
		str += fmt.Sprintf(" - (%v:%v -> %v:%v)",
			srcdests.Local.Host, srcdests.Local.Port,
			srcdests.Remote.Host, srcdests.Remote.Port,
		)
	}
	return str
}

func (tunnel *SSHtunnel) Close() {
	tunnel.client.Close()
	for key, listener := range tunnel.listener {
		if q, ok := tunnel.quit[key]; ok {
			close(q)
		}
		listener.Close()
	}
	tunnel.wg.Wait()
}

func (tunnel *SSHtunnel) Start() error {
	var err error
	tunnel.log.Info("starting ssh connection listener")

	tunnel.log.Infof("dialing ssh: %v", tunnel.String())
	address, _ := url.Parse(fmt.Sprintf("ssh://%s", tunnel.server.String()))
	tunnel.pool.GetConnection(address, tunnel.config)
	tunnel.client, err = ssh.Dial("tcp", tunnel.server.String(), tunnel.config)
	if err != nil {
		return errors.Wrapf(err, "server dial error to %v", tunnel.server.String())
	}

	for key, t := range tunnel.tunnels {
		tunnel.listener[key], err = net.Listen("tcp", t.Local.String())
		if err != nil {
			return errors.Wrapf(err, "cannot start listener on %v", t.Local.String())
		}
		tunnel.quit[key] = make(chan interface{})

		go func(k string) {
			//defer tunnel.wg.Done()
			conn, err := tunnel.listener[k].Accept()
			if err != nil {
				select {
				case <-tunnel.quit[k]:
					return
				default:
					tunnel.log.Errorf("error accepting connection on %v", tunnel.tunnels[k].Local.String())
				}
			} else {
				tunnel.wg.Add(1)
				go func() {
					tunnel.forward(conn, tunnel.tunnels[k].Remote)
					tunnel.wg.Done()
				}()
			}
		}(key)
	}
	return nil
}

func (tunnel *SSHtunnel) forward(localConn net.Conn, endpoint *Endpoint) {
	var err error

	remoteConn, err := tunnel.client.Dial("tcp", endpoint.String())
	if err != nil {
		tunnel.log.Errorf("Remote dial error %v: %v", endpoint.String(), err)
		return
	}

	copyConn := func(writer, reader net.Conn) {
		defer writer.Close()
		defer reader.Close()

		_, err := io.Copy(writer, reader)
		if err != nil {
			tunnel.log.Errorf("io.Copy error %v: %v", endpoint.String(), err)
		}
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		copyConn(localConn, remoteConn)
		wg.Done()
	}()
	go func() {
		copyConn(remoteConn, localConn)
		wg.Done()
	}()
	wg.Wait()
}
