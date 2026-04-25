package parking

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"

	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

type wtSession struct {
	session *webtransport.Session
}

func (s *wtSession) Send(data []byte) error {
	str, err := s.session.OpenStreamSync(context.Background())
	if err != nil {
		return err
	}
	defer str.Close()
	_, err = str.Write(data)
	return err
}

func (s *wtSession) Close() error {
	return s.session.CloseWithError(0, "")
}

type Server struct {
	hub      *Hub
	server   *webtransport.Server
	certFile string
	keyFile  string
}

func NewServer(hub *Hub, certFile, keyFile string) *Server {
	return &Server{
		hub:      hub,
		certFile: certFile,
		keyFile:  keyFile,
		server: &webtransport.Server{
			H3: &http3.Server{
				TLSConfig: &tls.Config{
					MinVersion: tls.VersionTLS13,
				},
			},
		},
	}
}

func (s *Server) Run(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/wt", s.handleUpgrade)

	s.server.H3.Addr = addr
	s.server.H3.Handler = mux

	return s.server.ListenAndServeTLS(s.certFile, s.keyFile)
}

func (s *Server) handleUpgrade(w http.ResponseWriter, r *http.Request) {
	sess, err := s.server.Upgrade(w, r)
	if err != nil {
		return
	}

	go func(sess *webtransport.Session) {
		ws := &wtSession{session: sess}
		s.hub.Add(ws)
		defer s.hub.Remove(ws)
		defer sess.CloseWithError(0, "")

		for {
			stream, err := sess.AcceptStream(context.Background())
			if err != nil {
				return
			}
			_, _ = io.Copy(io.Discard, stream)
			_ = stream.Close()
		}
	}(sess)
}
