package main

import (
	"github.com/y1015860449/gotoolkit/httpClient"
	"log"
)

func main() {
	cli := httpClient.NewHttp()
	rest, code, err := cli.Get("https://www.baidu.com", nil, nil)
	log.Printf("rest(%v) code(%v) err(%+v) \n", string(rest), code, err)
	rest, code, err = cli.Get("https://www.mob.com/mobService/mobpush?plat=0&product=MobGrow", nil, nil)
	log.Printf("rest(%v) code(%v) err(%+v) \n", string(rest), code, err)
	rest, code, err = cli.PostForm("https://fclog.baidu.com/log/ocpcagl", nil, map[string]string{"type": "behavior", "emd": "euc"})
	log.Printf("rest(%v) code(%v) err(%+v) \n", string(rest), code, err)
}
