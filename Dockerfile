FROM scratch

ENTRYPOINT ["/usr/bin/serverbin"]
CMD ["--help"]

COPY serverbin /usr/bin/serverbin