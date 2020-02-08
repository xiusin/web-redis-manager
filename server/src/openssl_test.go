package src

import (
  "fmt"
  "github.com/Luzifer/go-openssl"
  "testing"
)

func Test_ExampleOpenSSL_EncryptBytes(t *testing.T) {
  plaintext := "Hello World!"
  passphrase := "z4yH36a6zerhfE5427ZV"

  o := openssl.New()

  enc, err := o.EncryptBytes(passphrase, []byte(plaintext))
  if err != nil {
    fmt.Printf("An error occurred: %s\n", err)
  }

  fmt.Printf("Encrypted text: %s\n", string(enc))
}

func Test_ExampleOpenSSL_DecryptBytes(t *testing.T) {
  opensslEncrypted := "U2FsdGVkX1+5JiqOUtaMCsVQcqSUIaVR1spb7PBZQGE="
  passphrase := "z4yH36a6zerhfE5427ZV"

  o := openssl.New()

  dec, err := o.DecryptBytes(passphrase, []byte(opensslEncrypted))
  if err != nil {
    fmt.Printf("An error occurred: %s\n", err)
  }

  fmt.Printf("Decrypted text: %s\n", string(dec))

  // Output:
  // Decrypted text: hallowelt
}
