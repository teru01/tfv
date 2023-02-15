build:
	go build -o tfv ./cmd/tfv

clean:
	rm ./tfv

test:
	go test -v ./...
