GUINAME    = dollgui
DIR 	   = gui
LINKERFLAG = --Xcc="-I${DIR}"
PACKAGES   = --pkg gtk+-3.0 --pkg posix

EXENAME = dollwipe

build:
	@echo Building GUI...
	valac -o ${GUINAME} ${LINKERFLAG} ${DIR}/gui.vala ${DIR}/fails.vala ${DIR}/utils.vala ${DIR}/base.vala ${DIR}/consts.vala ${PACKAGES}
	@echo OK, ${GUINAME} built.
	@echo Building engine...
	/usr/bin/go/bin/go build -o ${EXENAME}
	@echo OK, ${EXENAME} built.

clean:
	rm ${GUINAME} ${EXENAME}
