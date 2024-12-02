FROM scratch

COPY mqtt-exporter /mqtt-exporter

ENTRYPOINT ["/mqtt-exporter"]
