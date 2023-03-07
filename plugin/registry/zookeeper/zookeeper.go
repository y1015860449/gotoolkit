package zookeeper

import (
	"github.com/go-zookeeper/zk"
	"time"
)

type ZkConfig struct {
	Urls    []string
	Timeout time.Duration
}

type ZkClient struct {
	conn   *zk.Conn
	events <-chan zk.Event
	conf   *ZkConfig
}

func InitZookeeper(c *ZkConfig) (*ZkClient, error) {
	conn, event, err := zk.Connect(c.Urls, c.Timeout)
	if err != nil {
		return nil, err
	}
	return &ZkClient{
		conn:   conn,
		events: event,
		conf:   nil,
	}, nil
}

func (cli *ZkClient) Close() {
	cli.conn.Close()
}

func (cli *ZkClient) Exists(path string) (bool, error) {
	exist, _, err := cli.conn.Exists(path)
	return exist, err
}

func (cli *ZkClient) Create(path string, flags int32, data []byte) error {
	acl := zk.WorldACL(zk.PermAll)
	_, err := cli.conn.Create(path, data, flags, acl)
	return err
}

func (cli *ZkClient) Modify(path string, data []byte) error {
	_, sate, _ := cli.conn.Get(path)
	_, err := cli.conn.Set(path, data, sate.Version)
	return err
}

func (cli *ZkClient) Delete(path string) error {
	_, sate, _ := cli.conn.Get(path)
	err := cli.conn.Delete(path, sate.Version)
	return err
}

func (cli *ZkClient) Get(path string) ([]byte, error) {
	data, _, err := cli.conn.Get(path)
	return data, err
}

func (cli *ZkClient) Children(path string) ([]string, error) {
	urls, _, err := cli.conn.Children(path)
	return urls, err
}

// 分布式事务锁
func (cli *ZkClient) Lock(path string) (*zk.Lock, error) {
	lock := zk.NewLock(cli.conn, path, zk.WorldACL(zk.PermAll))
	err := lock.Lock()
	return lock, err
}

func (cli *ZkClient) Unlock(lock *zk.Lock) error {
	return lock.Unlock()
}
