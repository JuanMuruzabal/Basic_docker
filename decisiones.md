
# JUSTIFICACIONES



# 1. ELEGIR Y PREPARAR APP

# eleccion de aplicacion (GO):
-Go genera un unico binario, sin dependencias externas
    Go tiene un compilador estático: cuando compilas con go build, el compilador incluye dentro del binario todas las librerías necesarias

-Es rapido y eficiente para desplegar en contenedores
    Como solo necesitás un binario, la imagen de Docker puede contener únicamente ese archivo → imágenes de apenas unos MB

-Permite imágenes Docker muy pequeñas, ideales para prácticas de DevOps.

--> Esto fue consultado a IA y con conocimientos previos de arqui de soft I y II del porque usamos Go en la materia

- la app tambien fue generada con IA

# Entorno Docker en Windows
- Instalé Docker Desktop for Windows, que incluye Docker Engine y Docker CLI.
- Docker Desktop utiliza WSL2 (Windows Subsystem for Linux 2) como backend, lo que permite ejecutar contenedores Linux de manera eficiente en Windows.
  
- Verifiqué la instalación con:
  -powershell
   docker --version
   docker run hello-world

Esta eleccion fue por los siguientes motivos:

Permite ejecutar contenedores Linux en Windows de forma transparente.
Se integra con WSL2, facilitando el uso de comandos desde PowerShell, CMD o directamente en una terminal Linux.
Es el estándar recomendado para desarrollo y pruebas en Windows.


# 2. CONSTRUIR UNA IMAGEN PERSONALIZADA

# Dockerfile

WORKDIR /app

- Copiar go.mod y go.sum para descargar dependencias
COPY go.mod ./
RUN go mod download

- Copiar el resto del código
COPY . .

- Usé un multi-stage build:
  
  La primera etapa (`golang:1.24-bookworm`) compila el código con las últimas actualizaciones de seguridad --> descargue la version
  La segunda etapa (`gcr.io/distroless/base-debian12`) contiene únicamente el binario ya compilado, reduciendo la superficie de ataque y el tamaño de la imagen.
   --> "Distroless" no incluye shell, gestor de paquetes ni herramientas extra: solo lo mínimo para correr el binario → mucho más seguro y liviano. Ya que quitamos la posibilidad de ejecutar comandos de shell

- Se fuerza un binario estatico con `CGO_ENABLED=0` y flags de optimización (`-ldflags="-w -s"`) para hacerlo más ligero y seguro.
- Se ejecuta con el usuario `nonroot`, evitando riesgos asociados a correr procesos como root dentro del contenedor.
- Se expone el puerto `8080`, donde el servidor HTTP escucha las conexiones.
- El comando por defecto es ejecutar el binario `./miapp`.


--> Con esta configuración se logra:
- Una imagen muy pequeña (aprox 6-10 MB frente a > 20 MB con Debian slim).
- Mayor seguridad (sin shell, sin gestor de paquetes, sin herramientas innecesarias).
- Ejecución con permisos mínimos gracias al usuario no root.
- Un despliegue más rápido y confiable en cualquier entorno Docker.


ESTAS CONFIGURACIONES FUERON LOGRADAS AL OPTIMIZAR CON IA UN DOCKER FILE CREADO POR MI --> ya que el codigo me describia que existian vulnerabilidades con ayuda de la IA puede actualizarlo de esta forma

Dockerfile anterior: 


    FROM golang:1.22 AS builder --> version generica y no especifica

    WORKDIR /app


    COPY go.mod ./
    RUN go mod download


    COPY . .


    RUN go build -o miapp . --> Binario normal


    FROM debian:bullseye-slim --> Incluye shell, gestor de paquetes y herramientas mínimas. 22MB

    WORKDIR /root/


    COPY --from=builder /app/miapp .

    (no se especifica el nonroot)

    EXPOSE 8080

    CMD ["./miapp"]

- La imagen se buildeo y se etiqueto como lo dice el README


# 3 PUBLICAR LA IMAGEN EN DOCKER HUB

  - ejecutamos docker login --> nos logeamos

  - como ya tenemos la imagen con tag la subimos

  - subimos la imagen --> docker push juanmuruzabal/app:v1.0


    Uso tags semánticos en este caso la version v1.0 para identificar la evolución de la aplicación.

    Mantengo un tag latest que apunta a la versión más reciente y estable:

    - juanmuruzabal/app-go:latest → última versión estable.

    - juanmuruzabal/app:v1.0 → primera versión publicada.

    De esta forma otros usuarios pueden elegir:

    Descargar siempre la última versión o fijar una versión específica para garantizar reproducibilidad en despliegues.

# 4. INTEGRAR UNA BASE DE DATOS AL CONTENEDOR (DOCKER CLI)

- Eleccion de base de datos

    elegi Mysql 8.0 ya que es con lo que trabajamos a lo largo de la carrera y ya tengo 2 años de experiencia con la misma, es apliamente usada en proyectos web y gracias a las librerias anteriores tiene muy buena integracion con GO.

- Levantar SQL
  
        docker run -d --name mysql --network app-net ^
        -e MYSQL_ROOT_PASSWORD=rootpass ^
        -e MYSQL_DATABASE=appdb ^
        -e MYSQL_USER=appuser ^
        -e MYSQL_PASSWORD=apppass ^
        -v mysql_data:/var/lib/mysql ^
        -v "%cd%\init.sql:/docker-entrypoint-initdb.d/init.sql" ^
        mysql:8.0 --default-authentication-plugin=mysql_native_password

    Este fue el comando utilizado para crear la imagen de mysql 8.0 donde: 

    --> Se utilizó una red personalizada (app-net) para permitir la comunicación entre los contenedores.
    --> Se montó un volumen persistente (mysql_data) para asegurar la conservación de los datos.
    --> Además, se montó un script init.sql en el contenedor de MySQL para inicializar la base y el usuario de aplicación.

        -- Create database if not exists
        CREATE DATABASE IF NOT EXISTS appdb;

        -- Create application user with proper authentication (with existence check)
        DROP USER IF EXISTS 'appuser'@'%';
        CREATE USER IF NOT EXISTS 'appuser'@'%' IDENTIFIED WITH mysql_native_password BY 'apppass';
        GRANT ALL PRIVILEGES ON appdb.* TO 'appuser'@'%';
        FLUSH PRIVILEGES;

        Esta estrategia evita errores de permisos y facilita que la aplicación se conecte correctamente sin usar el usuario root. --> error que estaba teniendo, la solucion fue sacada con IA

- Levantar la app

    docker run -d --name app --network app-net -p 8080:8080 ^
    -e DB_HOST=mysql ^
    -e DB_USER=appuser ^
    -e DB_PASSWORD=apppass ^
    -e DB_NAME=appdb ^
    juanmuruzabal/app:v1.1

    En esta parte creamos antes de ejecutar este programa la imagen juanmuruzabal/app:v1.1, que contiene la app con la logica de la base de datos

    LA APP FUE GENERADA CON IA --> La app mediante endpoints muestra un numero, se guarda el numero 42 esn la base de datos y con el endpoint /numero se muestra el numero almacenado.


 # 5. CONFIGURACION DE QA Y PROD

 --> utilizamos la ultima imagen juanmuruzabal/app:v1.1 para utilizarla en QA Y PROD donde:

- QA: Base de datos de pruebas, logs más verbosos, debug habilitado.
- PROD: Base de datos real, logs mínimos, sin debug.

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

--> Ambos usan la misma imagen v1.1
--> Los puertos externos cambian (8081 y 8082) para no chocar


La diferenciación entre QA y PROD se realiza a través de variables de entorno, lo que permite:
- Evitar duplicación de imágenes y asegurar que el código desplegado es exactamente el mismo.
- Facilitar la trazabilidad: cualquier bug encontrado en QA se puede reproducir en PROD con la misma imagen.
- Mejorar la portabilidad: al cambiar solo las variables (APP_ENV, DB_HOST, DB_USER, etc.) se puede adaptar la aplicación a diferentes   entornos sin modificar el código ni reconstruir la imagen.
- Se definió la variable APP_ENV para activar modos de log (verbose en QA, minimal en PROD) y se ajustaron las credenciales de conexión a base de datos mediante DB_HOST, DB_USER, DB_PASSWORD y DB_NAME. --> utilizada IA


# 6. DOCKER COMPOSE
   
   version: '3.8'

services:

  qa:
    build: 
      context: .
    image: juanmuruzabal/app:v1.2
    ports:
      - "8081:8080"
    environment:
      APP_ENV: QA
      DB_HOST: mysql
      DB_USER: appuser
      DB_PASSWORD: apppass
      DB_NAME: appdb_qa
    depends_on:
      mysql:
        condition: service_healthy

  prod:
    build:
      context: .
    image: juanmuruzabal/app:v1.2
    ports:
      - "8082:8080"
    environment:
      APP_ENV: PROD
      DB_HOST: mysql
      DB_USER: appuser
      DB_PASSWORD: apppass
      DB_NAME: appdb_prod
    depends_on:
      mysql:
        condition: service_healthy


  mysql:
    image: mysql:8.0
    command: sh -c "chown -R mysql:mysql /var/lib/mysql && exec docker-entrypoint.sh mysqld"
    user: mysql
    restart: unless-stopped
    volumes:
      - mysql_data:/var/lib/mysql
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    environment:
      MYSQL_ROOT_PASSWORD: rootpass
      MYSQL_USER: appuser
      MYSQL_PASSWORD: apppass
      MYSQL_INITDB_SKIP_TZINFO: 1
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-uroot", "-prootpass"]
      interval: 15s
      timeout: 30s
      retries: 10
      start_period: 120s

volumes:
  mysql_data:

  
- init.sql --> propuesto por IA

-- Create database if not exists
CREATE DATABASE IF NOT EXISTS appdb_qa;
CREATE DATABASE IF NOT EXISTS appdb_prod;

-- Create application user with proper authentication (with existence check)
DROP USER IF EXISTS 'appuser'@'%';
CREATE USER IF NOT EXISTS 'appuser'@'%' IDENTIFIED WITH mysql_native_password BY 'apppass';

GRANT ALL PRIVILEGES ON appdb_qa.* TO 'appuser'@'%';
GRANT ALL PRIVILEGES ON appdb_prod.* TO 'appuser'@'%';
FLUSH PRIVILEGES;

  
- Justificación del docker-compose

1. Definición de servicios --> qa, prod y mysql

   - Se definen dos entornos separados qa y prod que utilizan la misma imagen juanmuruzabal/app:v1.2, pero con variables de entorno distintas (APP_ENV, DB_NAME) para que cada instancia apunte a su base de datos correspondiente.
   Esto permite levantar en paralelo ambos entornos con una única configuración centralizada.

   - Control de arranque con depends_on y healthcheck -->
   Cada servicio de la aplicación depende de que MySQL esté realmente listo.
   Gracias al healthcheck, el contenedor de la aplicación no intentará conectarse hasta que MySQL haya inicializado correctamente, evitando fallos de conexión al iniciar.

   - Versión de la imagen v1.2 -->
   Se crea una nueva versión de la aplicación que incluye la capacidad de mostrar el entorno en el que se está ejecutando en QA o PROD --> Esto aporta trazabilidad y facilita las pruebas.

2. Persistencia de datos con volúmenes

    - Se define un volumen app_mysql_data montado en /var/lib/mysql, que es la ruta donde MySQL almacena sus archivos internos de datos --> se genera con app al principio ya que docker compose toma el nombre de la carpeta del proyecto como prefijo.

    - Esto asegura persistencia: aunque el contenedor MySQL se borre o reinicie, los datos de QA y PROD permanecen en el host --> a diferencia del otro volumen.

    - Se eligió un único contenedor MySQL con múltiples bases (appdb_qa y appdb_prod), lo cual simplifica la gestión, reduce el consumo de recursos y pero mantiene el aislamiento lógico entre entornos.

        --> a diferencia del ejercicio anterior, que prod y qa compartian el mismo volumen de datos pero sin separacion logica en las tablas, respondian a diferentes puertos pero no habia aislamiento logico.
   
3. Uso del archivo init.sql

    - Crea las bases de datos necesarias (appdb_qa y appdb_prod).

    - Crea el usuario appuser con contraseña apppass.

    - Se otorgan permisos específicos sobre ambas bases, evitando que se use el usuario root para la conexión de la aplicación.

    - Se fuerza el plugin mysql_native_password, resolviendo problemas de compatibilidad con algunos drivers (por ejemplo, de Go o Python) que no soportan caching_sha2_password de MySQL 8.
  
    --->Eso te asegura que cualquier persona que levante tu stack en otra máquina tendrá la misma base mínima para que la app funcione --> pero no es escalable, se buscara no depender del archivo despues del arranque inicial

4. Networks

 --> los contenedores ahora corren sobre la "app_default" ya que en el docker compose no se especifico la red, asi que se crea una por deaful

5. command: sh -c "chown -R mysql:mysql /var/lib/mysql && exec docker-entrypoint.sh mysqld"

    Motivo: Almacenar datos en un volumen externo puede generar conflictos de permisos entre el sistema de archivos del host y el usuario que ejecuta MySQL dentro del contenedor.

    Con este comando se fuerza a que todo el directorio /var/lib/mysql quede bajo propiedad del usuario mysql antes de iniciar el daemon mysqld.

    Luego, exec docker-entrypoint.sh mysqld arranca el servidor normalmente con permisos corregidos.

    - Esto resuelve errores comunes de inicio donde MySQL no puede acceder a sus propios archivos de datos por problemas de ownership.
   
    --> sacado con IA, tenia un problema de conflictos de permisos

6. VENTAJAS
   
   - Entornos reproducibles: QA y PROD comparten la misma imagen pero están desacoplados por variables de entorno y bases de datos distintas.

   - Persistencia: los datos sobreviven a reinicios gracias al volumen mysql_data.

   - Automatización: la inicialización automática con init.sql garantiza coherencia en la creación de usuarios y permisos.

   - Resiliencia: la aplicación solo intenta conectarse a la base cuando esta está sana (healthcheck).

   - Mantenimiento simplificado: un solo contenedor MySQL con múltiples bases resulta más ligero que mantener múltiples instancias, pero sigue permitiendo separación lógica de entornos.
  

  --> TODO EL PUNTO DE DOCKER COMPOSE FUE APOYADO POR IA YA QUE POSEIA DIVSOS ERRORES