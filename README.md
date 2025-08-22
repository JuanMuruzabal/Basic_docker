## INSTRUCCIONES

# CONSTRUCCION DE IMAGENES

- Para construir la imagen primero  descargamos la version 1.24  de go (la mas nueva, mas segura)
  
 docker pull golang:1.24-bookworm 

- Buildeamos la imagen con: 
  
  - creamos el dockerfile:
- 
        FROM golang:1.24-bookworm AS builder
        WORKDIR /app

        COPY go.mod ./
        RUN go mod download

        COPY . .

        RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o miapp .

        FROM gcr.io/distroless/base-debian12:nonroot

        WORKDIR /app

        COPY --from=builder --chown=nonroot:nonroot /app/miapp .

        USER nonroot

        EXPOSE 8080

        CMD ["./miapp"]


  luego hacemos docker build -t juanmuruzabal/app . 
  
  con el nombre de dockerHUB para el desarrollo de la app el . significa que el docker busque en la carpeta actual para crearla imagen

- Etiquetamos la imagen:
  docker tag juanmuruzabal/app juanmuruzabal/app:v1.0 --> ultima etiqueta de desarrollo segun versiones

  y mantenemos la lastest para version estable de desarrollo

# MONTAR VOLUMEN DE DATOS PERSISTENTES

docker run -d --name mysql --network app-net ^
        -e MYSQL_ROOT_PASSWORD=rootpass ^
        -e MYSQL_DATABASE=appdb ^
        -e MYSQL_USER=appuser ^
        -e MYSQL_PASSWORD=apppass ^
        -v mysql_data:/var/lib/mysql ^
        -v "%cd%\init.sql:/docker-entrypoint-initdb.d/init.sql" ^
        mysql:8.0 --default-authentication-plugin=mysql_native_password

# NOTAS IMPORTANTES

--> - DEBERA TENER EL INIT EN LA CARPETA APP

--> - Como se cambio el programa para que usara logica de datos se creo la imagen juanmuruzabal/app:v1.1 que es la que se usara

--> - Los comandos estan configurados para ponerlos en la CMD de windows

--> - Todos los contenedores van a ser parte de la red app-net en adelante

# LEVANTAR APP

    docker run -d --name app --network app-net -p 8080:8080 ^
    -e DB_HOST=mysql ^
    -e DB_USER=appuser ^
    -e DB_PASSWORD=apppass ^
    -e DB_NAME=appdb ^
    juanmuruzabal/app:v1.1

# ENDPOINTS A PROBAR 

http://localhost:8080/numero 
http://localhost:8080/incrementar


# CONFIGURACION DE QA Y PROD


- CREAR CONTENEDOR QA

docker run -d --name app-qa --network app-net -p 8081:8080 ^
-e APP_ENV=qa ^
-e DB_HOST=mysql ^
-e DB_USER=appuser ^
-e DB_PASSWORD=apppass ^
-e DB_NAME=appdb ^
juanmuruzabal/app:v1.1


- CREAR CONTENEDOR PROD

docker run -d --name app-prod --network app-net -p 8082:8080 ^
-e APP_ENV=prod ^
-e DB_HOST=mysql ^
-e DB_USER=appuser ^
-e DB_PASSWORD=apppass ^
-e DB_NAME=appdb ^
juanmuruzabal/app:v1.1


# DOCKER COMPOSE 

docker-compose build --no-cache --> para evitar errores y se creara la imagen v1.2 con la actualizacion en el codigo
docker-compose up

-probar

http://localhost:8081/numero 
http://localhost:8081/incrementar
http://localhost:8082/numero 
http://localhost:8082/incrementar
http://localhost:8081/ambiente
http://localhost:8082/ambiente

# PUBLICACION DE LOS CAMBIOS FINALES (QA)

- suponemos que la aplicacion (la imagen tomada en el docker compose) cumplio los estandares del entorno PROD

cambiamos la etiqueta para subir docker tag juanmuruzabal/app:v1.2 juanmuruzabal/app:lastest

donde lastest es la imagen de produccion (por el flujo de trabajo) y la ultima

- pusheamos la imagen docker push juanmuruzabal/app:lastestv1.2 

EL ENTORNO LASTEST REMPLAZA EL QA

subimos tambien una version de produccion si se quiere

juanmuruzabal/app:v1.2 juanmuruzabal/app:PRODv1.2
docker push juanmuruzabal/app:PRODv1.2
