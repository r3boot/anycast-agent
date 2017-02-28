BUILD_DIR = ./build
CMD_DIR = ./cmd

AACTL = ${BUILD_DIR}/aactl
AACTL_SRC =  ${CMD_DIR}/aactl/aactl.go

ANYCAST_AGENT = ${BUILD_DIR}/anycast-agent
ANYCAST_AGENT_SRC = ${CMD_DIR}/anycast-agent/anycast-agent.go

all: ${AACTL} ${ANYCAST_AGENT}

${BUILD_DIR}:
	mkdir -p ${BUILD_DIR}

${AACTL}:
	go build -v -o ${AACTL} ${AACTL_SRC}

${ANYCAST_AGENT}:
	go build -v -o ${ANYCAST_AGENT} ${ANYCAST_AGENT_SRC}

clean:
	rm -rf ${BUILD_DIR} || true
