prefix = /usr/local
bindir = $(prefix)/bin
sharedir = $(prefix)/share
mandir = $(sharedir)/man
man1dir = $(mandir)/man1
export PATH := $(PATH):/usr/local/go/bin

all: gowebserver

gowebserver:
	@go build gowebserver.go

clean:
	@rm -f gowebserver

install: all
	@install gowebserver $(DESTDIR)$(bindir)
	@install -m 0644 gowebserver.1 $(DESTDIR)$(man1dir)
