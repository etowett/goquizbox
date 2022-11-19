package util

import (
	"crypto/md5"
	"encoding/hex"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

const (
	digits           = "0123456789"
	hexLetters       = "abcdef"
	lowerCaseLetters = "abcdefghijklmnopqrstuvwxyz"
	upperCaseLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func GenerateApiKeySecret() string {
	return generateRandomString(digits+lowerCaseLetters+upperCaseLetters, 40)
}

func GenerateApiKeyUsername() string {
	return generateRandomString(digits+lowerCaseLetters+upperCaseLetters, 20)
}

func GenerateApiKeyPassword() string {
	return generateRandomString(digits+lowerCaseLetters+upperCaseLetters, 40)
}

func GenerateEmailActivationKey() string {
	return generateRandomString(digits+lowerCaseLetters+upperCaseLetters, 5)
}

func GenerateEmailConfirmationKey() string {
	return generateRandomString(digits+lowerCaseLetters+upperCaseLetters, 16)
}

func GeneratePhoneActivationKey() string {
	return generateRandomString(digits, 5)
}

func GenerateSalt() string {
	return generateRandomString(digits+lowerCaseLetters+upperCaseLetters, 12)
}

func GenerateUUID() string {
	return uuid.New().String()
}

func GenerateAccountUUID() string {
	return generateRandomString(hexLetters+digits, 8)
}

func GenerateRandomString(length int) string {
	return generateRandomString(lowerCaseLetters, length)
}

func GenerateRandomNumber(length int) string {
	return generateRandomString(digits, length)
}

func GenerateAvatarFileNameUUID() string {
	return generateRandomString(upperCaseLetters+lowerCaseLetters+digits, 8)
}

func generateRandomString(charset string, length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func GenerateRandomIntSlice(inputLength, desiredLength int) []int {

	if inputLength < 1 {
		return []int{}
	}

	if desiredLength > inputLength {
		desiredLength = inputLength
	}

	intSlice := GenerateIntSequence(0, inputLength, 1)

	randomIntSlice := make([]int, desiredLength)
	pos := 0

	rand.Seed(time.Now().Unix())
	for range randomIntSlice {

		guess := rand.Intn(len(intSlice))

		randomIntSlice[pos] = intSlice[guess]
		pos++

		intSlice = append(intSlice[:guess], intSlice[guess+1:]...)
	}

	return randomIntSlice
}

func GenerateIntSequence(begin, end, step int) []int {

	if step < 1 {
		return []int{}
	}

	length := (end - begin) / step

	if length < 1 {
		length = length * -1
	}

	if (end-begin)%step != 0 {
		length++
	}

	ints := make([]int, length)
	pos := 0

	if begin < end {

		for i := begin; i < end; i += step {
			ints[pos] = i
			pos++
		}

	} else {

		for i := begin; i > end; i -= step {
			ints[pos] = i
			pos++
		}

	}

	return ints
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
