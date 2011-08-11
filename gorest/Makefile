include $(GOROOT)/src/Make.inc

TARG=gorest.googlecode.com/hg/gorest

GOFILES=\
    doc.go\
	api.go\
	gorest.go\
	mime.go\
	parse.go\
	reflect.go\
	marshaller.go\
	client.go\
	util.go\


GOFMT=gofmt
BADFMT:=$(shell $(GOFMT) -l $(GOFILES) $(CGOFILES) $(wildcard *_test.go) 2> /dev/null)

gofmt: $(BADFMT)
	@for F in $(BADFMT); do $(GOFMT) -w $$F && echo $$F; done

ifneq ($(BADFMT),)
ifneq ($(MAKECMDGOALS),gofmt)
$(warning WARNING: make gofmt: $(BADFMT))
endif
endif


include $(GOROOT)/src/Make.pkg
