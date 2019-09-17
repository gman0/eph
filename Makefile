all: eph

ifeq ($(PREFIX),)
    PREFIX := /usr/local
endif

eph:
	mkdir -p _output
	go build -o _output/eph main.go

install: eph
	install -d $(DESTDIR)$(PREFIX)/bin/
	install -m 755 eph $(DESTDIR)$(PREFIX)/bin/

clean:
	rm -rf _output

.PHONY: eph all install clean
