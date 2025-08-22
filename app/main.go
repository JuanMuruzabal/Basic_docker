package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Modelo en la base de datos
type Config struct {
	ID    uint `gorm:"primaryKey"`
	Clave string
	Valor int
}

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// cadena de conexión a MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Error conectando a MySQL:", err)
	}

	// migrar modelo
	db.AutoMigrate(&Config{})

	// insertar un valor inicial si no existe
	var count int64
	db.Model(&Config{}).Where("clave = ?", "contador").Count(&count)
	if count == 0 {
		db.Create(&Config{Clave: "contador", Valor: 42})
	}

	// endpoint para traer el número
	http.HandleFunc("/ambiente", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Entorno: %s\n", os.Getenv("APP_ENV"))
	})

	http.HandleFunc("/numero", func(w http.ResponseWriter, r *http.Request) {
		var cfg Config
		if err := db.First(&cfg, "clave = ?", "contador").Error; err != nil {
			http.Error(w, "No se encontró el número", http.StatusNotFound)
			return
		}
		fmt.Fprintf(w, "El número guardado es: %d\n", cfg.Valor)
	})

	// endpoint para incrementar el número
	http.HandleFunc("/incrementar", func(w http.ResponseWriter, r *http.Request) {
		var cfg Config
		if err := db.First(&cfg, "clave = ?", "contador").Error; err != nil {
			http.Error(w, "No se encontró el contador", http.StatusNotFound)
			return
		}
		cfg.Valor++
		db.Save(&cfg)
		fmt.Fprintf(w, "El nuevo número es: %d\n", cfg.Valor)
	})

	log.Println("Servidor escuchando en :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
