package replacer

import (
	"crypto/sha1"
	"encoding/base32"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GehirnInc/crypt/apr1_crypt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/drone/envsubst"
	"golang.org/x/crypto/bcrypt"
	"regexp"
	"strings"
)

type replacer func(string) string

var pattern = regexp.MustCompile("::SECRET:([^:]+):SECRET::")

func SetPattern(newPattern string) error {
	p, err := regexp.Compile(newPattern)
	if err != nil {
		return err
	}
	pattern = p
	return nil
}

func ReplaceAll(str string) (string, error) {
	errs := map[string]error{}
	r, err := newAwsReplacer(func(replacement string, err error) {
		errs[replacement] = err
	})
	if err != nil {
		return str, err
	}
	res := replaceSecrets(str, r)

	if len(errs) > 0 {
		err = errors.New(fmt.Sprintf("found errors in replacements: %+v", errs))
	}

	return res, err
}

func replaceSecrets(str string, r replacer) string {
	return pattern.ReplaceAllStringFunc(str, func(s string) string {
		pattern := s[9 : len(s)-9]
		patternSubst, err := envsubst.EvalEnv(pattern)
		if err == nil {
			pattern = patternSubst
		}
		return r(pattern)
	})
}
func newAwsReplacer(errCallback func(string, error)) (replacer, error) {
	newSession, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	svc := secretsmanager.New(newSession)

	return func(secret string) string {
		secretId := secret
		key := ""
		encode := ""
		if strings.Contains(secret, "|") {
			splits := strings.Split(secret, "|")
			secretId = splits[0]
			key = splits[1]
			if len(splits) > 2 {
				encode = splits[2]
			}
		}

		input := &secretsmanager.GetSecretValueInput{
			SecretId: aws.String(secretId),
		}

		result, err := svc.GetSecretValue(input)
		if err != nil {
			errCallback(secret, err)
			return err.Error()
		}

		if key != "" {
			var values map[string]string
			err := json.Unmarshal([]byte(*result.SecretString), &values)
			if err != nil {
				errCallback(secret, err)
				return err.Error()
			}
			s, err := encodeValue(values[key], encode)
			if err != nil {
				errCallback(secret, err)
				return err.Error()
			}
			return s

		}

		s, err := encodeValue(*result.SecretString, encode)
		if err != nil {
			errCallback(secret, err)
			return err.Error()
		}
		return s
	}, nil
}

func encodeValue(value string, encode string) (string, error) {
	e := strings.ToLower(encode)
	if e == "base64" {
		return base64.StdEncoding.EncodeToString([]byte(value)), nil
	}
	if e == "base32" {
		return base32.StdEncoding.EncodeToString([]byte(value)), nil
	}
	if e == "apr1" {
		return apr1_crypt.New().Generate([]byte(value), nil)
	}
	if e == "bcrypt" {
		return toStringErr(bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost))
	}
	if e == "sha1" {
		s := sha1.New()
		s.Write([]byte(value))
		val := s.Sum(nil)
		return base64.StdEncoding.EncodeToString(val), nil
	}
	if e == "binary" {
		val := make([]byte, 0)
		_, err := base64.StdEncoding.Decode([]byte(value), val)
		return string(val), err
	}
	return value, nil
}

func toStringErr(i []byte, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return string(i), nil
}
