package proxy

import (
	"github.com/stretchr/testify/assert"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func FuzzGetProxy(f *testing.F) {
	f.Add(uint32(1), uint32(10))
	f.Fuzz(func(t *testing.T, index uint32, urlCounts uint32) {

		r := roundRobinSwitcher{}
		r.index = index
		r.proxyURLs = make([]*url.URL, urlCounts)

		for i := 0; i < int(urlCounts); i++ {
			r.proxyURLs[i] = &url.URL{}
			r.proxyURLs[i].Host = strconv.Itoa(i)
		}

		p, err := r.GetProxy(nil)
		if err != nil && strings.Contains(err.Error(), "empty proxy urls") {
			t.Skip()
		}

		assert.Nil(t, err)

		e := r.proxyURLs[index%urlCounts]

		if !reflect.DeepEqual(p, e) {
			t.Fail()
		}
	})
}
