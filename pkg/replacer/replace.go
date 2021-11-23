package replacer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
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
		return r(s[9 : len(s)-9])
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
		if strings.Contains(secret, "|") {
			splits := strings.Split(secret, "|")
			secretId = splits[0]
			key = splits[1]
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
			return values[key]
		}
		return *result.SecretString
	}, nil

}
