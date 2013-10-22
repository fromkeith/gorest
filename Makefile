include $(GOROOT)/src/Make.inc

TARG=code.google.com/p/gorest

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
	sec.go\



include $(GOROOT)/src/Make.pkg
