package exporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/danterobles/recolector/internal/collector"
)

// HTTPExporter serializa el SystemInfo como JSON y lo envía mediante HTTP POST
// a un endpoint configurable. Soporta autenticación Bearer token opcional.
//
// El cliente HTTP tiene un timeout de 10 segundos por defecto para evitar que
// una red lenta o un servidor no responsivo bloquee el ciclo de recolección.
// Este valor puede ser sobreescrito mediante WithHTTPTimeout.
type HTTPExporter struct {
	endpoint string       // URL completa del endpoint receptor (ej. http://servidor:8080/collect)
	token    string       // Bearer token para el header Authorization; vacío si no se requiere
	client   *http.Client // cliente HTTP inyectable para pruebas con httptest.Server
	pretty   bool         // si es true, envía JSON indentado
}

// HTTPOption es el tipo funcional para configurar un HTTPExporter.
type HTTPOption func(*HTTPExporter)

// WithHTTPClient reemplaza el http.Client del exportador.
// Usar en pruebas para inyectar el cliente del httptest.Server y evitar
// llamadas reales a la red durante la ejecución del test suite.
func WithHTTPClient(c *http.Client) HTTPOption {
	return func(e *HTTPExporter) {
		e.client = c
	}
}

// WithHTTPTimeout crea un nuevo http.Client con el timeout especificado
// y lo asigna al exportador. Útil para ajustar el timeout según las
// condiciones de red del entorno de producción.
func WithHTTPTimeout(d time.Duration) HTTPOption {
	return func(e *HTTPExporter) {
		e.client = &http.Client{Timeout: d}
	}
}

// NewHTTPExporter construye un HTTPExporter con un http.Client de 10 segundos de timeout.
// Si token es una cadena vacía, el header Authorization no se incluye en las peticiones.
func NewHTTPExporter(endpoint, token string, pretty bool, opts ...HTTPOption) *HTTPExporter {
	e := &HTTPExporter{
		endpoint: endpoint,
		token:    token,
		pretty:   pretty,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Export serializa info a JSON y realiza un HTTP POST al endpoint configurado.
//
// El proceso es:
//  1. Serializar SystemInfo a JSON (compacto o indentado según la configuración).
//  2. Construir la petición POST con Content-Type: application/json.
//  3. Añadir Authorization: Bearer <token> si el token no está vacío.
//  4. Ejecutar la petición con el cliente HTTP.
//  5. Drenar y cerrar el cuerpo de la respuesta para permitir la reutilización
//     de la conexión TCP (keep-alive).
//  6. Retornar error si el servidor responde con código HTTP >= 400.
func (e *HTTPExporter) Export(ctx context.Context, info *collector.SystemInfo) error {
	var (
		data []byte
		err  error
	)

	if e.pretty {
		data, err = json.MarshalIndent(info, "", "  ")
	} else {
		data, err = json.Marshal(info)
	}
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.endpoint, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if e.token != "" {
		req.Header.Set("Authorization", "Bearer "+e.token)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return fmt.Errorf("post to %s: %w", e.endpoint, err)
	}
	// Drenar el cuerpo antes de cerrar para que el runtime de Go pueda
	// reutilizar la conexión TCP subyacente en el siguiente ciclo de recolección.
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("el servidor respondió con %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return nil
}
