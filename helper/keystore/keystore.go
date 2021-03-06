package keystore

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
)

type createFn func() ([]byte, error)

// CreateIfNotExists generates a private key at the specified path,
// or reads the file on that path if it is present
func CreateIfNotExists(path string, create createFn) ([]byte, error) {
	_, err := os.Stat(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to stat (%s): %v", path, err)
	}

	var keyBuff []byte
	if !os.IsNotExist(err) {
		// Key exists
		
		keyBuff, err = ioutil.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("unable to read private key from disk (%s), %v", path, err)
		}

		return keyBuff, nil
	}

	// Key doesn't exist yet, generate it
	keyBuff, err = create()
	if err != nil {
		return nil, fmt.Errorf("unable to generate private key, %v", err)
	}

	// Encode it to a readable format (Base64) and write to disk
	keyBuff = []byte(hex.EncodeToString(keyBuff))
	if err = ioutil.WriteFile(path, keyBuff, 0600); err != nil {
		return nil, fmt.Errorf("unable to write private key to disk (%s), %v", path, err)
	}

	return keyBuff, nil
}
