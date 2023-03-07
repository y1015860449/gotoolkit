package consul

import (
	"fmt"
	"github.com/y1015860449/gotoolkit/log/zaplog"
	"net/http"
	"testing"
)

var count int64

func consulCheck(w http.ResponseWriter, r *http.Request) {

	s := "consulCheck" + fmt.Sprint(count) + "remote:" + r.RemoteAddr + " " + r.URL.String()
	fmt.Println(s)
	fmt.Fprintln(w, s)
	count++
}

func TestConsul(t *testing.T) {
	zaplog.InitLogger(nil)
	cli, err := InitConsul(&ConsulConfig{Address: "192.168.20.99:8500"})
	if err != nil {
		zaplog.ZapLog.Fatalf("err(%+v)", err)
	}
	//conf := &RegisterConfig{
	//	SvcId:   utils.GetUUID(),
	//	SvcName: "testConsul",
	//	Address: "192.168.166.125",
	//	Port:    8988,
	//	health: &HealthConfig{
	//		HealthType:                     "http",
	//		HealthUri:                      "http://192.168.166.125:8988/check",
	//		CheckID:                        utils.GetUUID(),
	//		Interval:                       "2s",
	//		Timeout:                        "1s",
	//		DeregisterCriticalServiceAfter: "3s",
	//	},
	//}
	//if err = cli.ServiceRegister(conf); err != nil {
	//	zaplog.ZapLog.Fatalf("err(%+v)", err)
	//}
	//defer cli.ServiceDeregister()

	list := []string{
		"81d53e62e03546179d09187207e9e4fa",
		"8f0983b2fea14214bf3242e7c3c5593e",
		"a239f746ecff46bf8515dd9b1c0058b1",
		"6797dcb3fdbb4f1bbfb2b0d3f4551692",
		"838d71ca258a4e0f85488f2736cf267a",
		"e25304844cc54b1eba22452217506f1d",
	}
	for _, id := range list {
		cli.Deregister(id)
	}

	//go func() {
	//	for {
	//		list, _ := cli.ServiceList("testConsul")
	//		zaplog.ZapLog.Infof("list(%+v)", list)
	//		time.Sleep(1000 * time.Millisecond)
	//	}
	//
	//}()
	//http.HandleFunc("/check", consulCheck)
	//http.ListenAndServe(fmt.Sprintf(":%d", 8988), nil)
}
