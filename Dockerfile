# syntax=docker/dockerfile:1

##
## Build
##
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS build

ARG TARGETOS
ARG TARGETARCH
ARG NAME
ARG VERSION
ARG REVISION
ARG ADDITIONAL_BUILD_PARAMS
ARG SKIP_LICENSES_REPORT=false
ARG VERSION=unknown
ARG REVISION=unknown

WORKDIR /app

RUN apk add --no-cache make
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
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
FROM azul/zulu-openjdk-alpine:26

ARG VERSION=unknown
ARG REVISION=unknown

LABEL "steadybit.com.discovery-disabled"="true"
LABEL "version"="${VERSION}"
LABEL "revision"="${REVISION}"
RUN echo "$VERSION" > /version.txt && echo "$REVISION" > /revision.txt

ENV MIRROR=https://downloads.apache.org/jmeter/binaries
ENV JMETER_VERSION=5.6.3
ENV JMETER_HOME=/opt/apache-jmeter-${JMETER_VERSION}
ENV JMETER_BIN=${JMETER_HOME}/bin
ENV PATH=${JMETER_BIN}:$PATH

## Installing dependencies
RUN apk add --update --no-cache coreutils bash procps wget && apk upgrade --no-cache

# Installing jmeter
ADD ${MIRROR}/apache-jmeter-${JMETER_VERSION}.tgz /tmp/
ADD ${MIRROR}/apache-jmeter-${JMETER_VERSION}.tgz.sha512 /tmp/
RUN mkdir -p /opt/ && cd /tmp/ \
 && sha512sum -c apache-jmeter-${JMETER_VERSION}.tgz.sha512 \
 && tar x -z -f apache-jmeter-${JMETER_VERSION}.tgz -C /opt \
 && rm -R -f apache* \
 && rm --recursive --force  ${JMETER_HOME}/docs \
 && chmod +x ${JMETER_HOME}/bin/*.sh

# Fix CVE-2025-48924: Replace vulnerable commons-lang3 with fixed version - remove after jmeter version with updated commons-lang3 is released
RUN cd /tmp \
 && rm -f ${JMETER_HOME}/lib/commons-lang3-*.jar \
 && wget --secure-protocol=TLSv1_2 --max-redirect=0 -q https://repo1.maven.org/maven2/org/apache/commons/commons-lang3/3.18.0/commons-lang3-3.18.0.jar -O ${JMETER_HOME}/lib/commons-lang3-3.18.0.jar

# Fix jackson 2.16.1 CVEs (e.g. CVE-2026-110879): replace with jackson 2.22 (as on jmeter master) - remove after jmeter version with updated jackson is released
RUN cd ${JMETER_HOME}/lib \
 && rm -f jackson-annotations-*.jar jackson-core-*.jar jackson-databind-*.jar \
 && wget --secure-protocol=TLSv1_2 --max-redirect=0 -q \
    https://repo1.maven.org/maven2/com/fasterxml/jackson/core/jackson-annotations/2.22/jackson-annotations-2.22.jar \
    https://repo1.maven.org/maven2/com/fasterxml/jackson/core/jackson-core/2.22.4/jackson-core-2.22.4.jar \
    https://repo1.maven.org/maven2/com/fasterxml/jackson/core/jackson-databind/2.22.4/jackson-databind-2.22.4.jar \
 && printf '%s\n' \
    "15c67f9498d6934cff9510fc70c5a12e52290457  jackson-annotations-2.22.jar" \
    "d753ae59bae7dcb183f856da541c5b27272088ce  jackson-core-2.22.4.jar" \
    "1ff5a55d62432f57e8044394976e598af0c5b3b8  jackson-databind-2.22.4.jar" \
  | sha1sum -c -

# Setup user
ARG USERNAME=steadybit
ARG USER_UID=10000
RUN adduser -u $USER_UID -D $USERNAME
USER $USER_UID

# Check installation
RUN jmeter --version

WORKDIR /

COPY --from=build /app/extension /extension
COPY --from=build /app/licenses /licenses

EXPOSE 8087
EXPOSE 8088

ENTRYPOINT ["/extension"]
