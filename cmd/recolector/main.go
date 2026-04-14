// Comando recolector: punto de entrada de la herramienta CLI.
//
// Responsabilidades de este paquete:
//  1. Parsear y validar los flags de la línea de comandos (delegado a internal/config).
//  2. Construir el Exporter adecuado según el modo de salida elegido.
//  3. Instanciar el SystemCollector.
//  4. Registrar los señales SIGINT y SIGTERM para un apagado limpio (graceful shutdown).
//  5. Ejecutar el ciclo de recolección mediante run(), que acepta interfaces y es
//     independientemente testeable sin realizar llamadas reales al sistema operativo.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danterobles/recolector/internal/collector"
	"github.com/danterobles/recolector/internal/config"
	"github.com/danterobles/recolector/internal/exporter"
)

func main() {
	// Parsear y validar los flags de la línea de comandos.
	cfg, err := config.FromFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "recolector: %v\n", err)
		os.Exit(1)
	}

	// Construir el exportador correspondiente al modo de salida elegido.
	exp, err := buildExporter(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "recolector: %v\n", err)
		os.Exit(1)
	}

	col := collector.New()

	// Crear un contexto que se cancela automáticamente al recibir SIGINT o SIGTERM.
	// Esto permite que un Ctrl-C o un kill detengan el loop de intervalo limpiamente.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Ejecutar el ciclo de recolección. context.Canceled no se trata como error
	// porque es el resultado esperado al detener el programa con Ctrl-C.
	if err := run(ctx, col, exp, cfg.Interval); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintf(os.Stderr, "recolector: %v\n", err)
		os.Exit(1)
	}
}

// buildExporter es una función de fábrica que devuelve el Exporter correcto
// según el modo de salida especificado en la configuración.
// Es una función pura (sin efectos secundarios) que facilita las pruebas unitarias.
func buildExporter(cfg config.Config) (exporter.Exporter, error) {
	switch cfg.Output {
	case "stdout", "":
		return exporter.NewStdoutExporter(cfg.Pretty), nil
	case "file":
		return exporter.NewFileExporter(cfg.FilePath, cfg.Pretty), nil
	case "post":
		return exporter.NewHTTPExporter(cfg.Endpoint, cfg.Token, cfg.Pretty), nil
	default:
		return nil, fmt.Errorf("modo de salida desconocido %q", cfg.Output)
	}
}

// run ejecuta el ciclo principal de recolección y exportación.
// Siempre realiza al menos una recolección inmediata al arrancar.
// Si interval es cero, termina tras esa primera recolección.
// Si interval es mayor que cero, repite la recolección en cada tick del timer
// hasta que el contexto sea cancelado (SIGINT/SIGTERM o timeout externo).
//
// Al recibir interfaces en lugar de tipos concretos, run puede ser probado
// con implementaciones mock sin necesidad de acceder al sistema operativo real.
func run(ctx context.Context, col collector.Collector, exp exporter.Exporter, interval time.Duration) error {
	// collect encapsula un ciclo completo: recolectar + exportar.
	collect := func() error {
		info, err := col.Collect(ctx)
		if err != nil {
			return fmt.Errorf("collect: %w", err)
		}
		if err := exp.Export(ctx, info); err != nil {
			return fmt.Errorf("export: %w", err)
		}
		return nil
	}

	// Primera ejecución inmediata, independiente del intervalo.
	if err := collect(); err != nil {
		return err
	}

	// Si no se especificó intervalo, terminar tras la primera recolección.
	if interval == 0 {
		return nil
	}

	// Loop de intervalo: esperar el siguiente tick o la cancelación del contexto.
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Contexto cancelado (Ctrl-C, SIGTERM, o timeout externo). Salida limpia.
			return ctx.Err()
		case <-ticker.C:
			if err := collect(); err != nil {
				return err
			}
		}
	}
}
