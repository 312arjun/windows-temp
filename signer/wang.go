package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"golang.org/x/crypto/blake2b"
)

// const (
// 	PubKeyStr   = "o1mgFxDznaTt4KfoDOhYCtrRd57s0Avg2PccpclQnUA="
// 	PubKeyExStr = "RWRhYmNkZWZnaKNZoBcQ852k7eCn6AzoWAra0Xee7NAL4Nj3HKXJUJ1A"
// 	PrivKeyStr  = "zH+yGwhPeLJl2bAC/oGQNObq5vYR0ErYmMZVS+FhcEqjWaAXEPOdpO3gp+gM6FgK2tF3nuzQC+DY9xylyVCdQA=="
// )

var (
	PubKeyStr   string
	PubKeyExStr string
	PrivKeyStr  string
)

func Ed25519Generate() {
	//GenerateKey will generate the private and public key pairs using
	//rand.Rander as source of entropy
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Fatal(err)
	}

	var pubKey2 [ed25519.PublicKeySize + 10]byte
	prefix := []byte("Edabcdefgh")
	for i := 0; i < ed25519.PublicKeySize+10; i++ {
		if i < 10 {
			pubKey2[i] = prefix[i]
		} else {
			pubKey2[i] = pubKey[i-10]
		}
	}

	PubKeyStr = base64.StdEncoding.EncodeToString(pubKey[:])
	PubKeyExStr = base64.StdEncoding.EncodeToString(pubKey2[:])
	PrivKeyStr = base64.StdEncoding.EncodeToString(privKey)

	// fmt.Println("Public Key: ", pubKeyStr)
	fmt.Println("Public Key: ", PubKeyExStr)
	// fmt.Println("Private Key: ", PrivKeyStr)

	// Write to a file
	f, err := os.Create("keys.bin")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	toWrite := fmt.Sprintf("%s\n%s", PubKeyExStr, PrivKeyStr)
	_, err2 := f.WriteString(toWrite)

	if err2 != nil {
		log.Fatal(err2)
	}

}

func SignAndVerify() {
	data := "abc123!?$*&()'-=@~"

	// Sign
	privKey, err := base64.StdEncoding.DecodeString(PrivKeyStr)
	if err != nil {
		log.Fatal(err)
	}

	signature := ed25519.Sign(privKey, []byte(data))

	// Verify
	pubKey, err := base64.StdEncoding.DecodeString(PubKeyStr)
	if err != nil {
		log.Fatal(err)
	}

	isValid := ed25519.Verify(pubKey, []byte(data), signature)
	if !isValid {
		fmt.Println("Invalid")
	}

	fmt.Println("Verified...")

}

func Base64Test() {
	data := "abc123!?$*&()'-=@~"
	encoded := "RWRNqGKtBXftKTKPpBPGDMe8jHLnFQ0EdRy8Wg0apV6vTDFLAODD83G4"

	sEnc := base64.StdEncoding.EncodeToString([]byte(data))
	fmt.Println(sEnc)

	sDec, _ := base64.StdEncoding.DecodeString(sEnc)
	fmt.Println(string(sDec))

	sDec2, _ := base64.StdEncoding.DecodeString(encoded)
	if sDec2[0] != 'E' || sDec2[1] != 'd' {
		fmt.Println("Invalid public key")
	}

}

func HashFiles() map[string]string {
	var count int

	files, err := ioutil.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}

	count = len(files)
	result := make(map[string]string, count)

	for _, file := range files {
		if !file.IsDir() && path.Ext(file.Name()) == ".msi" {
			data, err := ioutil.ReadFile(file.Name())
			if err != nil {
				log.Println("Couldn't read msi file")
				continue
			}

			// Calculate hash
			hash := blake2b.Sum256(data)
			hashStr := hex.EncodeToString(hash[:])

			result[hashStr] = file.Name()
		}
	}

	return result
}

func CreateSigFile(genKey bool) {
	if genKey {
		// Create ed25519 keys
		Ed25519Generate()
		return
	}

	// Hash files
	fileHashes := HashFiles()

	// Create  file to write to
	f, err := os.Create("latest.sig")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Concate the contents
	toWrite := ""
	for hash, fname := range fileHashes {
		toWrite += fmt.Sprintf("%s  %s\n", hash, fname)
	}

	// Sign the contents
	privKey, err := base64.StdEncoding.DecodeString(PrivKeyStr)
	if err != nil {
		log.Fatal(err)
	}
	signature := ed25519.Sign(privKey, []byte(toWrite))

	pubKey2, err := base64.StdEncoding.DecodeString(PubKeyExStr)
	if err != nil {
		log.Fatal(err)
	}

	var signature2 [ed25519.SignatureSize + 10]byte
	for i := 0; i < ed25519.SignatureSize+10; i++ {
		if i < 10 {
			signature2[i] = pubKey2[i]
		} else {
			signature2[i] = signature[i-10]
		}
	}

	toWrite = base64.StdEncoding.EncodeToString(signature2[:]) + "\n" + toWrite

	toWrite = "untrusted comment: verify with wireguard-windows-release.pubd" + "\n" + toWrite

	_, err2 := f.WriteString(toWrite)

	if err2 != nil {
		log.Fatal(err2)
	}
}

type fileList map[string][blake2b.Size256]byte

func VerifyTest() (fileList, error) {
	input, err := ioutil.ReadFile("latest.sig")
	if err != nil {
		log.Fatal(err)
	}

	publicKeyBytes, err := base64.StdEncoding.DecodeString(PubKeyExStr)
	if err != nil || len(publicKeyBytes) != ed25519.PublicKeySize+10 || publicKeyBytes[0] != 'E' || publicKeyBytes[1] != 'd' {
		return nil, errors.New("invalid public key")
	}
	lines := bytes.SplitN(input, []byte{'\n'}, 3)
	if len(lines) != 3 {
		return nil, errors.New("signature input has too few lines")
	}
	if !bytes.HasPrefix(lines[0], []byte("untrusted comment: ")) {
		return nil, errors.New("signature input is missing untrusted comment")
	}
	signatureBytes, err := base64.StdEncoding.DecodeString(string(lines[1]))
	if err != nil {
		return nil, errors.New("signature input is not valid base64")
	}
	if len(signatureBytes) != ed25519.SignatureSize+10 || !bytes.Equal(signatureBytes[:10], publicKeyBytes[:10]) {
		return nil, errors.New("signature input bytes are incorrect length, type, or keyid")
	}

	if !ed25519.Verify(publicKeyBytes[10:], lines[2], signatureBytes[10:]) {
		return nil, errors.New("signature is invalid")
	}

	fileLines := strings.Split(string(lines[2]), "\n")
	fileHashes := make(map[string][blake2b.Size256]byte, len(fileLines))

	return fileHashes, nil

}

func main() {
	if len(os.Args) == 2 && os.Args[1] == "/genkey" {
		CreateSigFile(true)
		return
	}

	// Read public key and private key from file
	data, err := ioutil.ReadFile("keys.bin")
	if err != nil {
		log.Fatal("couldn't read keys file")
	}

	fileLines := strings.Split(string(string(data)), "\n")
	PubKeyExStr = fileLines[0]
	PrivKeyStr = fileLines[1]

	// Create sig file
	CreateSigFile(false)

	_, err2 := VerifyTest()
	if err2 != nil {
		log.Printf("%+v", err2)
	}
}
