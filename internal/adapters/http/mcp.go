package http

import (
	"net/http"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/mark3labs/mcp-go/server"
)

// SetMCPTransport mounts the MCP Streamable HTTP transport at /mcp.
//
// mcp-go exposes a transport-agnostic Handle method so the MCP server can run
// inside Fiber without the net/http adaptor. The response writer reports that
// it cannot stream, so MCP replies are buffered as application/json. Wabasi's
// MCP surface is read-only request/response, so the optional SSE GET stream is
// not needed.
func (s *Server) SetMCPTransport(transport *server.StreamableHTTPServer) {
	s.app.All("/mcp", func(c *fiber.Ctx) error {
		transport.Handle(&fiberResponseWriter{ctx: c}, &server.HTTPRequest{
			Method:  c.Method(),
			URL:     requestURL(c),
			Header:  requestHeaders(c),
			Body:    c.Body(),
			Context: c.UserContext(),
		})
		return nil
	})
}

func requestURL(c *fiber.Ctx) *url.URL {
	u, err := url.ParseRequestURI(c.OriginalURL())
	if err != nil {
		return nil
	}
	return u
}

func requestHeaders(c *fiber.Ctx) http.Header {
	headers := make(http.Header)
	c.Request().Header.VisitAll(func(key, value []byte) {
		headers.Add(string(key), string(value))
	})
	return headers
}

type fiberResponseWriter struct {
	ctx         *fiber.Ctx
	header      http.Header
	wroteHeader bool
}

func (w *fiberResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *fiberResponseWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.ctx.Status(status)
	for key, values := range w.header {
		for _, value := range values {
			w.ctx.Set(key, value)
		}
	}
}

func (w *fiberResponseWriter) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ctx.Write(p)
}

func (w *fiberResponseWriter) Flush() {}

func (w *fiberResponseWriter) CanStream() bool { return false }
