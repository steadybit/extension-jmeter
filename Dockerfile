# syntax=docker/dockerfile:1

##
## Build
##
FROM golang:1.20-bullseye AS build

ARG NAME
ARG VERSION
ARG REVISION

WORKDIR /app

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y --no-install-recommends build-essential
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build \
    -ldflags="\
    -X 'github.com/steadybit/extension-kit/extbuild.ExtensionName=${NAME}' \
    -X 'github.com/steadybit/extension-kit/extbuild.Version=${VERSION}' \
    -X 'github.com/steadybit/extension-kit/extbuild.Revision=${REVISION}'" \
    -o ./extension

##
## Runtime
##
FROM openjdk:21-slim

ENV MIRROR https://www-eu.apache.org/dist/jmeter/binaries
ENV JMETER_VERSION 5.5
ENV JMETER_HOME /opt/apache-jmeter-${JMETER_VERSION}
ENV JMETER_BIN ${JMETER_HOME}/bin
ENV PATH ${JMETER_BIN}:$PATH

## Installing dependencies
RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y --no-install-recommends wget coreutils unzip bash curl procps

# Installing jmeter
RUN mkdir -p /opt/
ADD ${MIRROR}/apache-jmeter-${JMETER_VERSION}.tgz /tmp/
ADD ${MIRROR}/apache-jmeter-${JMETER_VERSION}.tgz.sha512 /tmp/
RUN cd /tmp/ \
 && sha512sum -c apache-jmeter-${JMETER_VERSION}.tgz.sha512 \
 && tar x -z -f apache-jmeter-${JMETER_VERSION}.tgz -C /opt \
 && rm -R -f apache* \
 && rm --recursive --force  ${JMETER_HOME}/docs \
 && chmod +x ${JMETER_HOME}/bin/*.sh

# Setup user
ARG USERNAME=steadybit
ARG USER_UID=10000
RUN adduser --uid $USER_UID $USERNAME
USER $USERNAME

# Check installation
RUN jmeter --version

WORKDIR /

COPY --from=build /app/extension /extension

EXPOSE 8087
EXPOSE 8088

ENTRYPOINT ["/extension"]
