<img src="./logo.png" height="130" align="right" alt="JMeter logo">

# Steadybit extension-jmeter

A [Steadybit](https://www.steadybit.com/) action implementation to integrate jmeter load tests into Steadybit experiments.

Learn about the capabilities of this extension in our [Reliability Hub](https://hub.steadybit.com/extension/com.steadybit.extension_jmeter).

## Configuration

The extension supports all environment variables provided by [steadybit/extension-kit](https://github.com/steadybit/extension-kit#environment-variables).

## Installation

### Kubernetes

Detailed information about agent and extension installation in kubernetes can also be found in
our [documentation](https://docs.steadybit.com/install-and-configure/install-agent/install-on-kubernetes).

#### Recommended (via agent helm chart)

All extensions provide a helm chart that is also integrated in the
[helm-chart](https://github.com/steadybit/helm-charts/tree/main/charts/steadybit-agent) of the agent.

You must provide additional values to activate this extension.

```
--set extension-jmeter.enabled=true \
```

Additional configuration options can be found in
the [helm-chart](https://github.com/steadybit/extension-jmeter/blob/main/charts/steadybit-extension-jmeter/values.yaml) of the
extension.

#### Alternative (via own helm chart)

If you need more control, you can install the extension via its
dedicated [helm-chart](https://github.com/steadybit/extension-jmeter/blob/main/charts/steadybit-extension-jmeter).

```bash
helm repo add steadybit-extension-jmeter https://steadybit.github.io/extension-jmeter
helm repo update
helm upgrade steadybit-extension-jmeter \
    --install \
    --wait \
    --timeout 5m0s \
    --create-namespace \
    --namespace steadybit-agent \
    steadybit-extension-jmeter/steadybit-extension-jmeter
```

### Linux Package

This extension is currently not available as a Linux package.

## Extension registration

Make sure that the extension is registered with the agent. In most cases this is done automatically. Please refer to
the [documentation](https://docs.steadybit.com/install-and-configure/install-agent/extension-discovery) for more
information about extension registration and how to verify.
