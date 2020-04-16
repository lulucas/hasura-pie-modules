package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gorilla/schema"
	"net/url"
	"sort"
	"strings"
)

type SignFunc func(interface{}) (string, error)

type SignOption struct {
	IgnoreKeys []string
	IgnoreEmpty bool
	KeyValueFunc func(key, value string) string
	PostSignStrHook func(string) string
	JoinSep string
	HashFunc func(string) string
}

func NewSignFunc(opt SignOption) SignFunc {
	if opt.KeyValueFunc == nil {
		opt.KeyValueFunc = func(key, value string) string {
			return fmt.Sprintf("%s=%s", key, value)
		}
	}

	if opt.JoinSep == "" {
		opt.JoinSep = "&"
	}

	if opt.HashFunc == nil {
		opt.HashFunc = func(s string) string {
			hash := md5.New()
			hash.Write([]byte(s))

			return hex.EncodeToString(hash.Sum(nil))
		}
	}

	var encoder = schema.NewEncoder()
	return func(i interface{}) (string, error) {
		values := url.Values{}
		if err := encoder.Encode(i, values); err != nil {
			return "", err
		}

		// url.Values
		var keys []string
		for k := range values {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var pairsToSign []string
		for _, k := range keys {
			if contains(opt.IgnoreKeys, k) {
				continue
			}
			if opt.IgnoreEmpty && values.Get(k) == "" {
				continue
			}
			pairsToSign = append(pairsToSign, opt.KeyValueFunc(k, values.Get(k)))
		}

		strToSign := strings.Join(pairsToSign, opt.JoinSep)

		if opt.PostSignStrHook != nil {
			strToSign = opt.PostSignStrHook(strToSign)
		}

		return opt.HashFunc(strToSign), nil
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
