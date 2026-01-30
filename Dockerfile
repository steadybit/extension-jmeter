# syntax=docker/dockerfile:1

##
## Build
##
FROM --platform=$BUILDPLATFORM golang:1.25-alpine AS build

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
FROM azul/zulu-openjdk-alpine:25

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
RUN apk add --no-cache coreutils bash procps wget

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
