package service

import (
	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"sync"
)

// Service ...
type Service struct {
	serv *machinery.Server
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
		_service = &Service{serv: server}
	})

	return _service
}

// NewWorker ...
func (s *Service) NewWorker() {
	worker := s.serv.NewWorker("work_conversion", 1)
	err := worker.Launch()
	if err != nil {
		// do something with the error
	}
}
