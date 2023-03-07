package hxzookeeper

import (
	"fmt"
	"github.com/go-zookeeper/zk"
)

type Register struct {
	conn    *zk.Conn
	conf    *ZkConfig
	svcInfo *ServiceInfo
}

func NewRegister(zkConfig *ZkConfig) (*Register, error) {
	conn, _, err := zk.Connect(zkConfig.Urls, zkConfig.Timeout)
	if err != nil {
		return nil, err
	}
	return &Register{
		conn: conn,
		conf: zkConfig,
	}, nil
}

func (r *Register) ServiceRegister(svcInfo *ServiceInfo) error {
	existsOrCreate := func(path string, flag int32) (bool, error) {
		exist, _, err := r.conn.Exists(path)
		if err != nil {
			return false, err
		}
		if !exist {
			_, err = r.conn.Create(path, nil, flag, zk.WorldACL(zk.PermAll))
			if err != nil {
				return false, err
			}
		}
		return exist, nil
	}

	node := fmt.Sprintf("/%s/%s", schemeName, svcInfo.svcName)
	if _, err := existsOrCreate(node, 0); err != nil {
		return err
	}
	path := fmt.Sprintf("/%s/%s/%s:%d", schemeName, svcInfo.svcName, svcInfo.SvcIp, svcInfo.SvcPort)
	exist, err := existsOrCreate(path, int32(zk.FlagEphemeral))
	if err != nil {
		return err
	}
	if exist {
		// 存在则更新
		if _, stat, err := r.conn.Get(path); err == nil {
			_, err = r.conn.Set(path, nil, stat.Version)
		}
	}
	return nil
}

func (r *Register) ServiceDeregister() error {
	path := fmt.Sprintf("%s/%s/%s:%d", schemeName, r.svcInfo.svcName, r.svcInfo.SvcIp, r.svcInfo.SvcPort)
	_, stat, err := r.conn.Get(path)
	if err != nil {
		return err
	}
	return r.conn.Delete(path, stat.Version)
}
