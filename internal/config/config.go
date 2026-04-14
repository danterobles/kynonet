// Package config gestiona la configuración de la herramienta a través de banderas (flags) de línea de comandos.
// Expone una única función pública, FromFlags, que parsea los argumentos y devuelve una Config validada.
// Al usar un flag.FlagSet local en lugar del global, el paquete es completamente testeable en aislamiento.
package config

import (
	"flag"
	"fmt"
	"net/url"
	"time"
)

// Config contiene todos los parámetros de ejecución de recolector.
// Se construye una sola vez al inicio del programa y se pasa por valor a los constructores.
type Config struct {
	// Output define el destino de los datos recolectados: "stdout", "file" o "post".
	Output string
	// FilePath es la ruta del archivo de salida; requerido cuando Output == "file".
	FilePath string
	// Endpoint es la URL del servidor receptor; requerido cuando Output == "post".
	Endpoint string
	// Token es el Bearer token para autenticación HTTP; puede estar vacío.
	Token string
	// Pretty activa el formato JSON indentado (2 espacios) en la salida.
	Pretty bool
	// Interval es el tiempo entre ciclos de recolección. Cero significa una sola ejecución.
	Interval time.Duration
}

// FromFlags parsea args (normalmente os.Args[1:]) con un FlagSet local
// y devuelve una Config validada. Retorna error si alguna combinación de
// banderas es inválida (por ejemplo, -output file sin -file).
func FromFlags(args []string) (Config, error) {
	fs := flag.NewFlagSet("recolector", flag.ContinueOnError)

	var (
		output   = fs.String("output", "stdout", `Destino de salida: stdout | file | post`)
		filePath = fs.String("file", "", "Ruta del archivo cuando -output=file")
		endpoint = fs.String("endpoint", "", "URL del endpoint cuando -output=post")
		token    = fs.String("token", "", "Bearer token para autenticación HTTP POST")
		pretty   = fs.Bool("pretty", false, "Formatear el JSON con indentación")
		interval = fs.Int("interval", 0, "Repetir la recolección cada N segundos (0 = una sola vez)")
	)

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	cfg := Config{
		Output:   *output,
		FilePath: *filePath,
		Endpoint: *endpoint,
		Token:    *token,
		Pretty:   *pretty,
		// Convierte segundos enteros a time.Duration para uso interno.
		Interval: time.Duration(*interval) * time.Second,
	}

	if err := cfg.validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// validate comprueba que la combinación de campos sea coherente.
// Las reglas son: "file" requiere FilePath no vacío; "post" requiere un Endpoint
// con URL válida; Interval no puede ser negativo.
func (c Config) validate() error {
	switch c.Output {
	case "stdout":
		// No se requiere ningún campo adicional.
	case "file":
		if c.FilePath == "" {
			return fmt.Errorf("output=file requiere -file <ruta>")
		}
	case "post":
		if c.Endpoint == "" {
			return fmt.Errorf("output=post requiere -endpoint <url>")
		}
		if _, err := url.ParseRequestURI(c.Endpoint); err != nil {
			return fmt.Errorf("URL de -endpoint inválida %q: %w", c.Endpoint, err)
		}
	default:
		return fmt.Errorf("modo de salida desconocido %q: debe ser stdout, file o post", c.Output)
	}

	if c.Interval < 0 {
		return fmt.Errorf("-interval debe ser >= 0")
	}

	return nil
}
