WindTurbine: server.go
	go build

.PHONY: run

run:
	go run server.go

.PHONY: clean

clean:
	rm WindTurbine
