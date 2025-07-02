# Stage 1: Build the frontend
FROM node:20-slim AS frontend
WORKDIR /app
RUN apt-get update && apt-get install -y git
RUN git clone https://github.com/stan-smith/OpenFLOW.git .
RUN npm install
RUN npm run build

# Stage 2: Build the Go application
FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY . .
COPY --from=frontend /app/build ./fossflow
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/fossflow-standalone .

# Stage 3: Final image
FROM scratch
WORKDIR /
COPY --from=builder /out/fossflow-standalone .
EXPOSE 8080
ENTRYPOINT ["./fossflow-standalone", "serve"]
