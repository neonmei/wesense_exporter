b:
  nix build

r:
  nix run

build:
  go build -o ./bin/ ./...

run: build
  ./bin/wesense_exporter

jaeger:
  # at http://127.0.0.1:16686/search
  podman run --rm --name jaeger -e COLLECTOR_ZIPKIN_HOST_PORT=:9411 -p 5775:5775/udp -p 6831:6831/udp -p 6832:6832/udp -p 5778:5778 -p 16686:16686 -p 14250:14250 -p 14268:14268 -p 14269:14269 -p 9411:9411 docker.io/jaegertracing/all-in-one:1.35

otel:
  otelcontribcol --config misc/opentelemetry-collector.yaml

prometheus:
  podman run --rm -v $PWD/misc:/etc/prometheus --name prometheus --network host prom/prometheus