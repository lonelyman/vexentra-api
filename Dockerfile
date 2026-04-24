FROM golang:1.25-alpine

RUN apk add --no-cache git build-base

WORKDIR /app

# Install tools once during image build (fast runtime)
RUN go install github.com/air-verse/air@latest && \
    go install github.com/pressly/goose/v3/cmd/goose@latest

ENV PATH="/go/bin:/usr/local/go/bin:${PATH}"

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# ตรวจสอบว่ามีไฟล์ .air.toml ก่อนรัน
CMD ["air", "-c", ".air.toml"]
