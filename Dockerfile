ARG GOLANG_VERSION=1.20

FROM golang:${GOLANG_VERSION} AS builder
ARG NODE_VERSION=14.17.6
RUN apt-get update && apt-get install -qq -y ca-certificates
RUN curl -O https://nodejs.org/dist/TODO_FILL_IN_THE_REST_OF_NODE_INSTALL_URL \
    && mkdir -p /usr/local/lib/nodejs \
    && tar -xzf TODO_NODE_FILE_PATH -C /usr/local/lib/nodejs \
    && rm TODO_NODE_FILE_PATH
ENV PATH=$PATH:/usr/local/lib/nodejs/TODO_NODE_FILE_PATH
ENV SERVICE_HOME /opt/weather-widget
WORKDIR ${SERVICE_HOME}
COPY ./cmd ./cmd
COPY ./web ./web
# COPY anything else that needs copied
RUN npm ci --prefix ./web
RUN npm run build --prefix ./web
COPY go.mod .
RUN CGO_ENABLED=0 go build -mod vendor -o ./out/weather-widget ./cmd/main
# or ./out/main ?? prob not ... 

FROM scratch
ARG SOME_SECRET
ENV SOME_SECRET ${SOME_SECRET}
ENV SERVICE_HOME /opt/weather-widget
ARG VERSION
ENV VERSION ${VERSION}
WORKDIR ${SERVICE_HOME}
COPY --from=builder ${SERVICE_HOME}/out/weather-widget .
COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY ./web ./web
COPY --from=builder ${SERVICE_HOME}/web/public/client ./web/public/client
EXPOSE 8080
ENTRYPOINT ["/opt/weather-widget/weather-widget"]