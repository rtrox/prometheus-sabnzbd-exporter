# [WIP] Prometheus Sabnzbd Exporter

Prometheus SabnzbD Exporter will, when polled by prometheus, collect statistics from SabnzbD API Endpoints.

![Grafana Dashboard](.github/images/dashboard.png)


Work in Progress! Will introduce README when ready for use. While the exporter functions, assume the flags, config, and metrics will be volatile and undocumented until you see a v0.0.1 release & official README.

## Operating the Exporter

Prometheus-SabnzbD-Exporter can be configured via flag, EnvVar, or Config File.
```bash
      --api_key string       api key of sabnzbd
      --base_url string      base url of sabnzbd
      --config strings       path to one or more .yaml config files
      --go_collector         enables go stats exporter
      --listen_port string   port to listen on (default "8080")
      --log_level string     log level (debug, info, warn, error) (default "info")
      --process_collector    enables process stats exporter
```

So normal usage would be:

```bash
prometheus-sabnzbd-exporter \
    --api_key <your key> \
    --base_url http://localhost:8080 \
    --listen_port 8081
```

## Running via Docker

```bash
docker run -d --name sabnzbd-exporter \
    -e SABNZBD_API_KEY=<yourkey>
    -e SABNZBD_BASE_URL=http://localhost:8080
    -e SABNZBD_LISTEN_PORT=8080
    ghcr.io/rtrox/prometheus-sabnzbd-exporter:latest
```

## Running via Kubernetes

Example Manifests are available in the [kubernetes/manifests](kubernetes/manifests) directory. Simply edit as needed and:

```bash
kubectl apply -f ./kubernetes/manifests/
```
