FROM golang:1.27.4-alpine

WORKDIR /app

# git necesario para go get de dependencias
RUN apk add --no-cache git build-base

# Redirigir cachés de Go a rutas que serán persistidas vía volúmenes nombrados
ENV GOCACHE=/root/.cache/go-build
ENV GOMODCACHE=/root/.cache/go-mod
# Instalar gopls (el motor de autocompletado, solo si se usa devcontainer)
RUN go install golang.org/x/tools/gopls@latest

CMD ["sleep", "infinity"]
