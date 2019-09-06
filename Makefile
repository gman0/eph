all: eph

eph:
	rm -f eph
	go build -o eph main.go
