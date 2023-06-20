<img src="./logo.png" height="130" align="right" alt="JMeter logo">

# Steadybit extension-jmeter

A [Steadybit](https://www.steadybit.com/) action implementation to integrate jmeter load tests into Steadybit experiments.

Learn about the capabilities of this extension in our [Reliability Hub](https://hub.steadybit.com/extension/com.github.steadybit.extension_jmeter).

## Configuration

The extension supports all environment variables provided by [steadybit/extension-kit](https://github.com/steadybit/extension-kit#environment-variables).

## Installation

### Using Docker

```sh
docker run \
  --rm \
  -p 8087 \
  --name steadybit-extension-jmeter \
  ghcr.io/steadybit/extension-jmeter:latest
```

### Using Helm in Kubernetes

```sh
helm repo add steadybit-extension-jmeter https://steadybit.github.io/extension-jmeter
helm repo update
helm upgrade steadybit-extension-jmeter \
    --install \
    --wait \
    --timeout 5m0s \
    --create-namespace \
    --namespace steadybit-extension \
    steadybit-extension-jmeter/steadybit-extension-jmeter
```

## Register the extension

Make sure to register the extension at the steadybit platform. Please refer to
the [documentation](https://docs.steadybit.com/integrate-with-steadybit/extensions/extension-installation) for more information.
