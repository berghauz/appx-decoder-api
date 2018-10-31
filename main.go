package main

func main() {

	c := createContext()
	logger.Infof("env set %+v", c)
	go c.runAPIServer()
	select {}
}
