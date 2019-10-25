# conversion
```go
        eng, err := conversion.InitMySQL(
            conversion.LoginOption(cfg.Addr, cfg.Username, cfg.Password))
    	if err != nil {
            return nil, fmt.Errorf("init conversion:%w", err)
    	}
        conversion.RegisterDatabase(eng)

        t := conversion.NewTask()
        t.Limit = cfg.Limit
        //don't stop when all task is done	
        t.SetAutoStop(false)        
        //then start
        t.Start()
```