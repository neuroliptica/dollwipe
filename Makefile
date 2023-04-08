GUINAME    = dollgui
EXENAME    = dollwipe
DIR 	   = gui
LINKERFLAG = --Xcc="-I${DIR}"
PACKAGES   = --pkg gtk+-3.0 --pkg posix

.PHONY: build gui engine

build: gui engine
	@echo Done

gui:
	@echo Building GUI...
	valac -o ${GUINAME} ${LINKERFLAG} ${DIR}/gui.vala ${DIR}/fails.vala ${DIR}/utils.vala ${DIR}/base.vala ${DIR}/consts.vala ${PACKAGES}
	@echo OK, ${GUINAME} built.

engine:
	@echo Building engine...
	go build -o ${EXENAME}
	@echo OK, ${EXENAME} built.

clean:
	rm ${GUINAME} ${EXENAME}
