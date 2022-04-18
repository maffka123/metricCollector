package handlers

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/netip"
	"strings"

	"github.com/maffka123/metricCollector/internal/server/config"
)

type Middleware func(http.Handler) http.HandlerFunc
type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

type metricUtils struct {
	cfg *config.Config
}

func newMetricUtils(cfg *config.Config) metricUtils {
	return metricUtils{cfg: cfg}
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func checkForLength(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := strings.Split(strings.Trim(r.URL.String(), "/"), "/")
		if len(q) < 3 {
			http.Error(w, "404 - No metric was given!", http.StatusNotFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func checkForPost(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func checkForGet(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func checkForJSON(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "This endpoint accepts only jsons", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func unpackGZIP(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, b := range r.Header.Values("Content-Encoding") {
			if b != "gzip" && b != "64base" {
				http.Error(w, "Only gzip encoding is allowed", http.StatusMethodNotAllowed)
				return
			}
		}
		if len(r.Header.Values("Content-Encoding")) == 0 ||
			(len(r.Header.Values("Content-Encoding")) == 1 && r.Header.Get("Content-Encoding") == "64base") {
			next.ServeHTTP(w, r)
			return
		}

		rw, err := gzip.NewReader(r.Body)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		r.Body = rw
		next.ServeHTTP(w, r)
	})
}

type rsaMW struct {
	key *rsa.PrivateKey
}

func NewRsaMW(key rsa.PrivateKey) rsaMW {
	return rsaMW{
		key: &key,
	}
}

func (rsaMW rsaMW) decodeRSA(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, b := range r.Header.Values("Content-Encoding") {
			if b != "gzip" && b != "64base" {
				http.Error(w, "Only 64base encoding is allowed", http.StatusMethodNotAllowed)
				return
			}
		}
		if len(r.Header.Values("Content-Encoding")) == 0 ||
			(len(r.Header.Values("Content-Encoding")) == 1 && r.Header.Get("Content-Encoding") == "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		hash := sha256.New()
		drb, err := rsa.DecryptOAEP(hash, rand.Reader, rsaMW.key, body, nil)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		r.Body = ioutil.NopCloser(bytes.NewBufferString(string(drb)))
		next.ServeHTTP(w, r)
	})
}

func packGZIP(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			io.WriteString(w, err.Error())
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

func Conveyor(h http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}

func (mu *metricUtils) checkTrusted(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if mu.cfg.TrustedSubnet != "" {
			b, err := IfIPinCIDR(mu.cfg.TrustedSubnet, r.Header.Get("X-Real-IP"))
			if err != nil {
				http.Error(w, fmt.Sprintf("IP address could not be parsed: %s", err), http.StatusForbidden)
				return
			}

			if !*b {
				http.Error(w, "IP address is not inside relible network", http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func IfIPinCIDR(cidr string, ipStr string) (*bool, error) {
	network, err := netip.ParsePrefix(cidr)
	if err != nil {
		return nil, err
	}

	ip, err := netip.ParseAddr(ipStr)
	if err != nil {
		return nil, err
	}

	b := network.Contains(ip)
	return &b, nil
}
