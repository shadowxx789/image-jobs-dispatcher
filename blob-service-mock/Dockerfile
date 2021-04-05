FROM theshamuel/baseimg-go-build:1.15.1 as builder

ARG VER
ARG SKIP_TESTS
ENV GOFLAGS="-mod=vendor"

RUN apk --no-cache add tzdata zip ca-certificates git

ADD . /build/blob-service-mock
ADD .golangci.yml /build/blob-service-mock/app/.golangci.yml
WORKDIR /build/blob-service-mock

#test
RUN \
    if [ -z "$SKIP_TESTS" ] ; then \
        go test -timeout=30s ./...; \
    else echo "[WARN] Skip tests" ; fi

#linter GolangCI
RUN \
    if [ -z "$SKIP_TESTS" ] ; then \
        golangci-lint run --skip-dirs vendor --config .golangci.yml ./...; \
    else echo "[WARN] Skip GolangCI linter" ; fi

RUN \
    version=$(date +%Y%m%d%H%M%S); \
    if [ -n "$VER" ] ; then \
    version=${VER}; fi; \
    echo "version=$version"; \
    go build -mod=vendor  -o blob-service-mock -ldflags "-X main.version=${version} -s -w" ./app

FROM theshamuel/baseimg-go-app:1.0-alpine3.10

WORKDIR /srv
COPY --from=builder /build/blob-service-mock/blob-service-mock /srv/blob-service-mock
COPY ./data /data

RUN chown -R appuser:appuser /srv
USER appuser

CMD [ "/srv/blob-service-mock" ]