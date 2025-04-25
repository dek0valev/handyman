FROM golang:1.24.2 AS build

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /handyman .

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian12

WORKDIR /

COPY --from=build /handyman /handyman

ENTRYPOINT ["/handyman"]
