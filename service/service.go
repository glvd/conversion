package service

import (
	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"sync"
)

// Service ...
type Service struct {
	*machinery.Server
}

var _service *Service
var _once = sync.Once{}

// NewService ...
func NewService() *Service {
	_once.Do(func() {
		var cfg = &config.Config{
			Broker:        "amqp://guest:guest@localhost:5672/",
			DefaultQueue:  "machinery_tasks",
			ResultBackend: "amqp://guest:guest@localhost:5672/",
			AMQP: &config.AMQPConfig{
				Exchange:     "machinery_exchange",
				ExchangeType: "direct",
				BindingKey:   "machinery_task",
			},
		}

		server, err := machinery.NewServer(cfg)
		if err != nil {
			return
		}
		_service = &Service{Server: server}
	})

	return _service
}
