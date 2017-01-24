package main

import (
	"github.com/stretchr/testify/assert"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

type TestData struct {
	MaxWorker           int
	MaxQueue            int
	ReqInterval         time.Duration
	ReqTimes            int
	ReqPreloadLen       int
	ReqPreloadLenFloat  int
	ReqPreloadWait      time.Duration
	ReqPreloadWaitFloat time.Duration
	GOMAXPROCS          int
}

func (td *TestData) Str() string {
	v := reflect.ValueOf(*td)
	t := v.Type()
	result := ""
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		result += fmt.Sprintf("%s:%v ", f.Name, v.Field(i).Interface())
	}
	return result
}

func getJsonData(t testing.TB, v interface{}) string {
	b, err := json.Marshal(v)
	assert.Equal(t, err, nil)
	return string(b)
}

func testSystem(td *TestData, t testing.TB) int {
	d := AsyncDispatcher(td.MaxWorker, td.MaxQueue)

	// make a test server
	sm := http.NewServeMux()
	sm.HandleFunc("/", newPayloadHandler(d.Queue, ReqMaxLength))

	for i := 0; i < td.ReqTimes; i++ {
		if d.IsBlock() {
			return i
		}
		rd := PayloadCollection{}
		loop := td.ReqPreloadLen
		if td.ReqPreloadLenFloat > 0 {
			loop += rand.Intn(td.ReqPreloadLenFloat)
		}
		for j := 0; j < loop; j++ {
			preload := Payload{td.ReqPreloadWait, 0, 0}
			if td.ReqPreloadWaitFloat > 0 {
				preload.Wait += time.Duration(rand.Int63n(int64(td.ReqPreloadWaitFloat)))
			}
			rd = append(rd, preload)
		}
		req, err := http.NewRequest("POST", "/", strings.NewReader(getJsonData(t, rd)))
		assert.Empty(t, err)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		sm.ServeHTTP(w, req)
		assert.Equal(t, w.Code, 200)
		time.Sleep(td.ReqInterval)
	}
	d.Stop()
	return td.ReqTimes
}

func TestRequest(t *testing.T) {
	log.SetOutput(os.Stderr)
	td := TestData{
		MaxWorker:           200,
		MaxQueue:            200,
		ReqInterval:         0,
		ReqTimes:            50,
		ReqPreloadLen:       5,
		ReqPreloadLenFloat:  0,
		ReqPreloadWait:      500 * time.Millisecond,
		ReqPreloadWaitFloat: 1000 * time.Millisecond,
		GOMAXPROCS:          runtime.GOMAXPROCS(0),
	}
	t.Logf("%s", td.Str())
	blockat := testSystem(&td, t)
	assert.Equal(t, blockat, td.ReqTimes)
}

func TestBlock(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	td := TestData{
		MaxWorker:           50,
		MaxQueue:            50,
		ReqInterval:         50 * time.Millisecond,
		ReqTimes:            50,
		ReqPreloadLen:       10,
		ReqPreloadLenFloat:  0,
		ReqPreloadWait:      50 * time.Millisecond,
		ReqPreloadWaitFloat: 0 * time.Millisecond,
		GOMAXPROCS:          runtime.GOMAXPROCS(0),
	}
	// incr 50 preload process time
	for i := 0; i < 5; i++ {
		td.ReqPreloadWait += 50 * time.Millisecond
		t.Logf("%s", td.Str())
		blockat := testSystem(&td, t)
		if blockat == td.ReqTimes {
			t.Log("not block")
		} else {
			t.Logf("block at %d th request", blockat)
			break
		}
	}
}
