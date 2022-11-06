# wesense_exporter

Hey internets! this repo is basically is a toy [OpenTelemetry] project that fetches metrics from the awesome [WeSense] sensor and exposes metrics about dioxide, relative humidity and temperature.

This should be used with a OTel endpoint which can make use of that metric or just use the OpenTelemetry [Collector] and make it export OpenMetrics/Prometheus metrics.

![Grafana Dashboard Example](misc/grafana.png)

## How to run?

If you got nix you're golden, and no dependencies gathering is necessary. Just make sure you have `direnv` and `nix-direnv` installed and do `direnv allow` to "sleeve into" a dev environment with dependencies installed (see `flake.nix`) when you cd into this repository.

When dependencies are up, adjust `.envrc` to make sure configuration is ok. Then simply run:

```bash
just run
```

Regarding configuring things with environment variables. For example, you might want to point the exporter to your sensor:

```
export WESENSE_EXPORTER_ENDPOINT="http://1.2.3.4:88"
```

Also, OpenTelemetry might be configured with [SDK variables](https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/protocol/exporter.md) like `OTEL_EXPORTER_OTLP_PROTOCOL`, and so on.

If you use nix, justyou might as well overlay the package and then do something like:

```nix
    systemd.user.services.wesense_exporter = {
      Unit.Description = "WeSense Exporter";
      Service.ExecStart = "${pkgs.wesense_exporter}/bin/wesense_exporter";
      Service.Environment = [
        "WESENSE_EXPORTER_ENDPOINT=http://1.2.3.4:88"
        "OTEL_EXPORTER_OTLP_INSECURE=true"
        "OTEL_EXPORTER_OTLP_ENDPOINT=http://127.0.0.1:4317"
        "OTEL_EXPORTER_OTLP_METRICS_ENDPOINT=http://127.0.0.1:4317"
        ];
    };
```

To auto-start the exporter.

[WeSense]: https://wesense.tech/
[OpenTelemetry]: https://opentelemetry.io/
[Collector]: https://opentelemetry.io/docs/collector/