package service

import (
	"github.com/RichardKnop/machinery/v1"
	"github.com/RichardKnop/machinery/v1/config"
	"sync"
)

// Service ...
type Service struct {
	Name string
	Task int
	err  chan error
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
			ResultBackend: "redis://localhost:6379",
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
		_service = &Service{
			Name: "work_conversion",
			Task: 1,
			err:  make(chan error),
			serv: server,
		}
	})

	return _service
}

// NewWorker ...
func (s *Service) NewWorker() {
	worker := s.serv.NewWorker(s.Name, s.Task)
	worker.LaunchAsync(s.err)
	return
}

// HandleWorker ...
func (s *Service) HandleWorker() error {
	return <-s.err
}

// AddTask ...
func (s *Service) AddTask(name string, val interface{}) error {
	return s.serv.RegisterTask(name, val)
}
