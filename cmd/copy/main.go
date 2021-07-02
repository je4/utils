package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"flag"
	"fmt"
	"github.com/blend/go-sdk/crypto"
	"github.com/gosuri/uiprogress"
	"github.com/je4/utils/v2/pkg/ssh"
	"github.com/je4/utils/v2/pkg/stream"
	"golang.org/x/crypto/sha3"
	"math"
	"net/url"
	"os"
	"regexp"
	"time"
)

const (
	loglevel             = "DEBUG"
	logFormat            = "%{time:2006-01-02T15:04:05.000} %{module}::%{shortfunc} [%{shortfile}] > %{level:.5s} - %{message}"
	maxClientConcurrency = 64
)

var targetRegex = regexp.MustCompile(`^([^@]+)@([^/:]+):([0-9]+)/(.+)$`)

func main() {
	var privateKey string
	flag.StringVar(&privateKey, "identity", "", "private key file")
	concurrency := flag.Int("concurrency", 50, "sftp client concurrency")
	maxPacketSize := flag.Int("maxpacketsize", 512*1024, "max packet size for sftp upload")
	flag.Parse()
	tail := flag.Args()
	if len(tail) < 2 {
		fmt.Println("invalid parameters")
		fmt.Printf("%s -identity [private key file] [sourcepath] [sftp target]\n", os.Args[0])
		os.Exit(1)
	}
	fi, err := os.Stat(privateKey)
	if err != nil {
		fmt.Printf("cannot stat %s: %v\n", privateKey, err)
		os.Exit(1)
	}
	if fi.IsDir() {
		fmt.Printf("%s is a directory\n", privateKey)
		os.Exit(1)
	}

	src := tail[0]
	fi, err = os.Stat(src)
	if err != nil {
		fmt.Printf("cannot stat %s: %v\n", src, err)
		os.Exit(1)
	}
	if fi.IsDir() {
		fmt.Printf("%s is a directory\n", src)
		os.Exit(1)
	}

	target := tail[1]
	matches := targetRegex.FindStringSubmatch(target)
	if matches == nil {
		fmt.Printf("invalid format for target %s\n", target)
		os.Exit(1)
	}
	targetUser := matches[1]
	targetHost := matches[2]
	targetPort := matches[3]
	targetPath := matches[4]

	logger, lf := CreateLogger("sftp", "", nil, loglevel, logFormat)
	defer lf.Close()

	/* ProgressReader Bar */
	uiprogress.Start()
	bar := uiprogress.AddBar(100)
	bar.AppendCompleted()
	bar.PrependElapsed()

	/* Encryption */
	key, err := crypto.CreateKey(crypto.DefaultKeySize)
	if err != nil {
		logger.Panicf("cannot generate crypto key: %v", err)
	}
	iv := make([]byte, aes.BlockSize)
	if _, err = rand.Read(iv); err != nil {
		logger.Panicf("cannot create random iv: %v", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		logger.Panicf("cannot create cipher block: %v", err)
	}
	ctrStream := cipher.NewCTR(block, iv)
	mac := hmac.New(sha256.New, key)

	/* ChecksumReaderWriter */
	hashSha512 := sha512.New()
	hashSha256 := sha256.New()
	hashSha3_224 := sha3.New224()
	hashSha3_384 := sha3.New384()

	rsc, err := stream.NewReadStreamQueue(
		stream.NewEncryptReader(block, ctrStream, mac, iv, logger),
		stream.NewProgressReaderWriter(
			fi.Size(),
			time.Second,
			func(remaining time.Duration, percent float64, estimated time.Time, complete bool) {
				bar.Set(int(math.Round(percent) + 1))
			}),
		stream.NewChecksumReaderWriter(hashSha512, logger),
		stream.NewChecksumReaderWriter(hashSha256, logger),
		stream.NewChecksumReaderWriter(hashSha3_224, logger),
		stream.NewChecksumReaderWriter(hashSha3_384, logger),
	)

	sftp, err := ssh.NewSFTP(
		[]string{privateKey}, "", "",
		*concurrency, maxClientConcurrency, *maxPacketSize,
		rsc,
		logger,
	)
	if err != nil {
		fmt.Printf("cannot initialize sftp: %v\n", err)
		os.Exit(1)
	}

	rawurl := fmt.Sprintf("sftp://%s@%s:%s/%s", targetUser, targetHost, targetPort, targetPath)
	targetUrl, err := url.Parse(rawurl)
	if err != nil {
		fmt.Printf("cannot parse url %s: %v\n", rawurl, err)
		os.Exit(1)
	}
	//	shaSink := sha512.New()
	//	dest := io.MultiWriter(w, shaSink)

	_, err = sftp.PutFile(targetUrl, src)
	if err != nil {
		fmt.Printf("cannot upload %s -> %s: %v\n", src, targetUrl.String(), err)
		os.Exit(1)
	}
	fmt.Printf("sha512: %x\n", hashSha512.Sum(nil))
	fmt.Printf("sha256: %x\n", hashSha256.Sum(nil))
	fmt.Printf("sha3_224: %x\n", hashSha3_224.Sum(nil))
	fmt.Printf("sha3_384: %x\n", hashSha3_384.Sum(nil))
	fmt.Printf("decrypt using openssl: \n openssl enc -aes-256-ctr -nosalt -d -in %s -out %s -K '%x' -iv '%x'\n", "encrypted.aes256", "plain.dat", key, iv)
	fmt.Printf("target: %s\n", targetUrl.String())
}
