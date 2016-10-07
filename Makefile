WindTurbine: structs.go server.go
	go build

.PHONY: run

run:
	go run structs.go server.go

.PHONY: clean

clean:
	rm WindTurbine
