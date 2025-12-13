package main

func main() {
	server := InitWebServer()
	server.Run(":8080")
}
