all: eph

ifeq ($(PREFIX),)
    PREFIX := /usr/local
endif

eph:
	rm -f eph
	go build -o eph main.go

install: eph
	install -d $(DESTDIR)$(PREFIX)/bin/
	install -m 755 eph $(DESTDIR)$(PREFIX)/bin/
