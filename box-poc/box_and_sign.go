package main

import (
	cryptRand "crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/sign"
)

const (
	encryptionKeyBasename = "crypt_curve25519"
	signingKeyBasename    = "sign_ed25519"
)

var (
	defaultKeyDirectory, _ = filepath.Abs(filepath.Join(".", "keys"))
)

type BoxAndSign struct {
	keyDirectory      string
	encryptionPubKey  *[32]byte
	encryptionPrivKey *[32]byte
	signingPubKey     *[32]byte
	signingPrivKey    *[64]byte
}

func NewBoxAndSign(keyDirectory string) (*BoxAndSign, error) {
	if "" == keyDirectory {
		keyDirectory = defaultKeyDirectory
	}
	keyDirectory, _ = filepath.Abs(keyDirectory)
	boxAndSign := &BoxAndSign{keyDirectory: keyDirectory}
	err := boxAndSign.create()
	fatalCheck(err)
	err = boxAndSign.store()
	return boxAndSign, err
}

func LoadBoxAndSign(keyDirectory string) (*BoxAndSign, error) {
	if "" == keyDirectory {
		keyDirectory = defaultKeyDirectory
	}
	keyDirectory, _ = filepath.Abs(keyDirectory)
	boxAndSign := &BoxAndSign{keyDirectory: keyDirectory}
	err := boxAndSign.read()
	return boxAndSign, err
}

func LoadOrCreateBoxAndSign(keyDirectory string) (*BoxAndSign, error) {
	if "" == keyDirectory {
		keyDirectory = defaultKeyDirectory
	}
	keyDirectory, _ = filepath.Abs(keyDirectory)
	boxAndSign := &BoxAndSign{keyDirectory: keyDirectory}
	err := boxAndSign.read()
	if nil != err {
		err := boxAndSign.create()
		fatalCheck(err)
		err = boxAndSign.store()
	}
	return boxAndSign, err
}

func (b *BoxAndSign) create() error {
	encryptPubKey, encryptPrivKey, err := box.GenerateKey(cryptRand.Reader)
	fatalCheck(err)
	signPubKey, signPrivKey, err := sign.GenerateKey(cryptRand.Reader)
	fatalCheck(err)
	b.encryptionPubKey = encryptPubKey
	b.encryptionPrivKey = encryptPrivKey
	b.signingPubKey = signPubKey
	b.signingPrivKey = signPrivKey
	return nil
}

func (b *BoxAndSign) store() error {
	err := os.MkdirAll(b.keyDirectory, 0655)
	fatalCheck(err)
	fileHandle, err := os.OpenFile(
		filepath.Join(b.keyDirectory, encryptionKeyBasename),
		os.O_CREATE|os.O_RDWR,
		0400,
	)
	fatalCheck(err)
	err = binary.Write(fileHandle, binary.BigEndian, b.encryptionPrivKey)
	fatalCheck(err)
	fatalCheck(fileHandle.Close())
	fileHandle, err = os.OpenFile(
		filepath.Join(b.keyDirectory, fmt.Sprintf("%s.pub", encryptionKeyBasename)),
		os.O_CREATE|os.O_RDWR,
		0644,
	)
	fatalCheck(err)
	err = binary.Write(fileHandle, binary.BigEndian, b.encryptionPubKey)
	fatalCheck(err)
	fatalCheck(fileHandle.Close())
	fileHandle, err = os.OpenFile(
		filepath.Join(b.keyDirectory, signingKeyBasename),
		os.O_CREATE|os.O_RDWR,
		0400,
	)
	fatalCheck(err)
	err = binary.Write(fileHandle, binary.BigEndian, b.signingPrivKey)
	fatalCheck(err)
	fatalCheck(fileHandle.Close())
	fileHandle, err = os.OpenFile(
		filepath.Join(b.keyDirectory, fmt.Sprintf("%s.pub", signingKeyBasename)),
		os.O_CREATE|os.O_RDWR,
		0644,
	)
	fatalCheck(err)
	err = binary.Write(fileHandle, binary.BigEndian, b.signingPubKey)
	fatalCheck(err)
	fatalCheck(fileHandle.Close())
	return nil
}

func (b *BoxAndSign) read() error {
	fileHandle, err := os.Open(filepath.Join(b.keyDirectory, encryptionKeyBasename))
	if nil != err {
		return err
	}
	b.encryptionPrivKey = new([32]byte)
	err = binary.Read(fileHandle, binary.BigEndian, b.encryptionPrivKey)
	fatalCheck(err)
	fatalCheck(fileHandle.Close())
	fileHandle, err = os.Open(filepath.Join(b.keyDirectory, fmt.Sprintf("%s.pub", encryptionKeyBasename)))
	if nil != err {
		return err
	}
	b.encryptionPubKey = new([32]byte)
	err = binary.Read(fileHandle, binary.BigEndian, b.encryptionPubKey)
	fatalCheck(err)
	fatalCheck(fileHandle.Close())
	fileHandle, err = os.Open(filepath.Join(b.keyDirectory, signingKeyBasename))
	if nil != err {
		return err
	}
	b.signingPrivKey = new([64]byte)
	err = binary.Read(fileHandle, binary.BigEndian, b.signingPrivKey)
	fatalCheck(err)
	fatalCheck(fileHandle.Close())
	fileHandle, err = os.Open(filepath.Join(b.keyDirectory, fmt.Sprintf("%s.pub", signingKeyBasename)))
	if nil != err {
		return err
	}
	b.signingPubKey = new([32]byte)
	err = binary.Read(fileHandle, binary.BigEndian, b.signingPubKey)
	fatalCheck(err)
	fatalCheck(fileHandle.Close())
	return nil
}

func (b *BoxAndSign) FreshNonce() [24]byte {
	var nonce [24]byte
	_, err := io.ReadFull(cryptRand.Reader, nonce[:])
	fatalCheck(err)
	return nonce
}

func (b *BoxAndSign) EncryptAndSign(message []byte, recipientEncryptPubKey *[32]byte) []byte {
	nonce := b.FreshNonce()
	encryptedMessage := box.Seal(nonce[:], message, &nonce, recipientEncryptPubKey, b.encryptionPrivKey)
	return sign.Sign(nil, encryptedMessage, b.signingPrivKey)
}

func (b *BoxAndSign) OpenAndDecrypt(signedMessage []byte, senderSignPubKey, senderEncryptPubKey *[32]byte) []byte {
	unsignedMessage, valid := sign.Open(nil, signedMessage, senderSignPubKey)
	if !valid {
		fatalCheck(fmt.Errorf("message was not signed: %b", unsignedMessage))
	}
	var decryptNonce [24]byte
	copy(decryptNonce[:], unsignedMessage[:24])
	message, ok := box.Open(nil, unsignedMessage[24:], &decryptNonce, senderEncryptPubKey, b.encryptionPrivKey)
	if !ok {
		fatalCheck(fmt.Errorf("decryption error: %b", message))
	}
	return message
}
