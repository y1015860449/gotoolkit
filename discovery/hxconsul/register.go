package hxconsul

import (
	"fmt"
	consulApi "github.com/hashicorp/consul/api"
	"github.com/y1015860449/gotoolkit/utils"
	"time"
)

type Register struct {
	client         *consulApi.Client
	consulConf     *ConsulConfig
	registerConf   *RegisterConfig
	svcId          string
	checkId        string
	deregisterChan chan bool
}

func NewRegister(consulConf *ConsulConfig) (*Register, error) {
	c := consulApi.DefaultConfig()
	c.Address = consulConf.Address
	cli, err := consulApi.NewClient(c)
	if err != nil {
		return nil, err
	}
	return &Register{
		client:         cli,
		consulConf:     consulConf,
		deregisterChan: make(chan bool, 1),
	}, nil
}

func (register *Register) ServiceRegister(registerConf *RegisterConfig) error {
	svcId := utils.GetUUID()
	reg := &consulApi.AgentServiceRegistration{
		ID:      svcId,
		Name:    registerConf.SvcName,
		Port:    registerConf.Port,
		Address: registerConf.Address,
	}
	if len(registerConf.Tag) > 0 {
		reg.Tags = []string{registerConf.Tag}
	}
	if err := register.client.Agent().ServiceRegister(reg); err != nil {
		return err
	}
	register.svcId = svcId
	if register.consulConf.Ttl <= 0 {
		register.consulConf.Ttl = 5
	}
	checkId := utils.GetUUID()
	check := consulApi.AgentServiceCheck{
		TTL:                            fmt.Sprintf("%ds", register.consulConf.Ttl),
		Status:                         consulApi.HealthPassing,
		Timeout:                        "1s",
		DeregisterCriticalServiceAfter: "5s",
	}
	if err := register.client.Agent().CheckRegister(&consulApi.AgentCheckRegistration{
		ID:                checkId,
		Name:              registerConf.SvcName,
		ServiceID:         svcId,
		AgentServiceCheck: check,
	}); err != nil {
		return err
	}
	register.checkId = checkId
	go func() {
		ticker := time.NewTicker(registerConf.UpdateInterval)
		for {
			select {
			case <-ticker.C:
				_ = register.client.Agent().UpdateTTL(checkId, "", check.Status)
			case <-register.deregisterChan:
				break
			}
		}
	}()
	return nil
}

func (register *Register) ServiceDeregister() error {
	register.deregisterChan <- true
	if err := register.client.Agent().ServiceDeregister(register.svcId); err != nil {
		return err
	}
	if err := register.client.Agent().CheckDeregister(register.checkId); err != nil {
		return err
	}
	return nil
}
