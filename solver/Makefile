SERVERNAME = solver

.PHONY: build clean

build:
	@echo Building solver server...
	go build -o ${SERVERNAME}
	@echo Ok, solver backend was built.
	@echo Use ./${SERVERNAME} to start local instance.

clean:
	rm ${SERVERNAME}
