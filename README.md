# helpers
## init log
	...
	h.LogInit(fmt.Sprintf("%s.log", os.Args[0]), 4095)
	defer h.LogRelease()
	...
	h.Info("info messsage")
	...
	logCustomFuncName := h.CustomLogFunc(h.CUSTOM, "FuncName")
	logCustomFuncName("start")
	logCustomFuncName2 := h.CustomLogFunc(h.CUSTOM<<1, "FuncName2")
	logCustomFuncName2("start second")
	...
## config
	...
	LOOP:
	for i, arg := range os.Args[1:] {
		switch {
		case arg == "-c" || arg == "--config":
			if i+2 < len(os.Args) {
				configPath := os.Args[i+2]
				configName := strings.Split(os.Args[0], "/")
				if err := h.ConfigInit(configPath, configName[len(configName)-1]); err != nil {
					panic(err)
				}
				break LOOP
			}
		case arg == "-h" || arg == "--help":
			fmt.Printf("Usage of %q:\n  -c/--config\t- config file\n  -h/--help\t- print this\n", os.Args[0])
			os.Exit(0)
		}
	}
	...
	someVar = h.CFG.String("some.var", "def value")
	...
