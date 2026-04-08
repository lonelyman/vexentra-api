FROM golang:1.25-alpine

RUN apk add --no-cache git build-base

WORKDIR /app

# ติดตั้ง air
RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# ตรวจสอบว่ามีไฟล์ .air.toml ก่อนรัน
CMD ["air", "-c", ".air.toml"]