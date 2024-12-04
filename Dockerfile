FROM golang:1.23 AS build

WORKDIR /app

COPY . /app

RUN CGO_ENABLED=1 go build -ldflags '-extldflags "-static"' -o app .

FROM alpine AS production

ENV APP_ENV=production

WORKDIR /app

# 为alpine创建时区
ENV TZ=Asia/Shanghai

COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=build /app/app /app/entrypoint /app/

COPY  ./config  /app/config

EXPOSE 8080

ENTRYPOINT [ "./entrypoint" ]

CMD ["run"]
