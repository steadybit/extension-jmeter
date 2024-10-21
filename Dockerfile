# syntax=docker/dockerfile:1

##
## Build
##
FROM --platform=$BUILDPLATFORM golang:1.23-bullseye AS build

ARG TARGETOS TARGETARCH
ARG NAME
ARG VERSION
ARG REVISION
ARG ADDITIONAL_BUILD_PARAMS
ARG SKIP_LICENSES_REPORT=false

WORKDIR /app

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y --no-install-recommends build-essential
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -ldflags="\
    -X 'github.com/steadybit/extension-kit/extbuild.ExtensionName=${NAME}' \
    -X 'github.com/steadybit/extension-kit/extbuild.Version=${VERSION}' \
    -X 'github.com/steadybit/extension-kit/extbuild.Revision=${REVISION}'" \
    -o ./extension \
    ${ADDITIONAL_BUILD_PARAMS}
RUN make licenses-report

##
## Runtime
##
FROM azul/zulu-openjdk-debian:21

LABEL "steadybit.com.discovery-disabled"="true"

ENV MIRROR=https://downloads.apache.org/jmeter/binaries
ENV JMETER_VERSION=5.6.3
ENV JMETER_HOME=/opt/apache-jmeter-${JMETER_VERSION}
ENV JMETER_BIN=${JMETER_HOME}/bin
ENV PATH=${JMETER_BIN}:$PATH

## Installing dependencies
RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y --no-install-recommends wget coreutils unzip bash curl procps

# Installing jmeter
ADD ${MIRROR}/apache-jmeter-${JMETER_VERSION}.tgz /tmp/
ADD ${MIRROR}/apache-jmeter-${JMETER_VERSION}.tgz.sha512 /tmp/
RUN mkdir -p /opt/ && cd /tmp/ \
 && sha512sum -c apache-jmeter-${JMETER_VERSION}.tgz.sha512 \
 && tar x -z -f apache-jmeter-${JMETER_VERSION}.tgz -C /opt \
 && rm -R -f apache* \
 && rm --recursive --force  ${JMETER_HOME}/docs \
 && chmod +x ${JMETER_HOME}/bin/*.sh

# Setup user
ARG USERNAME=steadybit
ARG USER_UID=10000
ARG USER_GID=$USER_UID
RUN groupadd --gid $USER_GID $USERNAME \
    && useradd --uid $USER_UID --gid $USER_GID -m $USERNAME
USER $USERNAME

# Check installation
RUN jmeter --version

WORKDIR /

COPY --from=build /app/extension /extension
COPY --from=build /app/licenses /licenses

EXPOSE 8087
EXPOSE 8088

ENTRYPOINT ["/extension"]
