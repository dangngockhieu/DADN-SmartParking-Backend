package parking

import (
	"backend/configs"
	"context"
	"crypto/tls"
	"encoding/binary"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

const (
	// writeTimeout là thời gian tối đa cho 1 lần write.
	// Client bị slow network quá timeout → disconnect.
	writeTimeout = 5 * time.Second
)

// wtSession wrap webtransport.Session với persistent unidirectional stream.
// Thay vì mở stream mới mỗi lần Send (overhead lớn), giữ 1 stream duy nhất
// và gửi length-prefixed frames: [4 bytes big-endian length][JSON payload].
type wtSession struct {
	session *webtransport.Session
	mu      sync.Mutex
	stream  *webtransport.SendStream
}

// ensureStream mở persistent stream nếu chưa có.
func (s *wtSession) ensureStream() error {
	if s.stream != nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()

	stream, err := s.session.OpenUniStreamSync(ctx)
	if err != nil {
		return err
	}

	s.stream = stream
	return nil
}

// Send gửi 1 message qua persistent stream với length-prefixed framing.
// Frame format: [4 bytes big-endian uint32 = len(data)][data bytes]
// Thread-safe nhờ mutex.
func (s *wtSession) Send(data []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureStream(); err != nil {
		return err
	}

	// Set write deadline để tránh block mãi nếu client bị lag
	_ = s.stream.SetWriteDeadline(time.Now().Add(writeTimeout))

	// Write 4-byte length header (big-endian)
	var header [4]byte
	binary.BigEndian.PutUint32(header[:], uint32(len(data)))
	if _, err := s.stream.Write(header[:]); err != nil {
		s.resetStream()
		return err
	}

	// Write payload
	if _, err := s.stream.Write(data); err != nil {
		s.resetStream()
		return err
	}

	return nil
}

// resetStream đóng stream hiện tại, lần Send tiếp theo sẽ mở stream mới.
func (s *wtSession) resetStream() {
	if s.stream != nil {
		_ = s.stream.Close()
		s.stream = nil
	}
}

// Close đóng toàn bộ session.
func (s *wtSession) Close() error {
	s.mu.Lock()
	s.resetStream()
	s.mu.Unlock()
	return s.session.CloseWithError(0, "")
}

// ─── Server ──────────────────────────────────────────────────────────────────

type Server struct {
	hub      *Hub
	server   *webtransport.Server
	certFile string
	keyFile  string
	cfg      *configs.Config
}

func NewServer(hub *Hub, certFile, keyFile string, cfg *configs.Config) *Server {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		panic(err)
	}

	h3Server := &http3.Server{
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS13,
			NextProtos:   []string{"h3"},
		},
	}

	// Bắt buộc phải gọi ConfigureHTTP3Server để bật WebTransport support
	// (enable datagrams, thêm SETTINGS enable_webtransport=1, setup ConnContext)
	webtransport.ConfigureHTTP3Server(h3Server)

	return &Server{
		hub:      hub,
		certFile: certFile,
		keyFile:  keyFile,
		cfg:      cfg,
		server: &webtransport.Server{
			H3: h3Server,
			CheckOrigin: func(r *http.Request) bool {
				origin := r.Header.Get("Origin")
				log.Println("[WT] Origin:", origin)

				return origin == cfg.FrontendURL
			},
		},
	}
}

func (s *Server) Run(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/wt", s.handleUpgrade)

	s.server.H3.Addr = addr
	s.server.H3.Handler = mux

	log.Printf("WebTransport HTTP/3 server listening on %s", addr)

	return s.server.ListenAndServe()
}

func (s *Server) handleUpgrade(w http.ResponseWriter, r *http.Request) {
	lotIDStr := r.URL.Query().Get("lotId")
	if lotIDStr == "" {
		log.Println("WebTransport reject: missing lotId")
		http.Error(w, "missing lotId", http.StatusBadRequest)
		return
	}

	lotID64, err := strconv.ParseUint(lotIDStr, 10, 64)
	if err != nil {
		log.Println("WebTransport reject: invalid lotId:", lotIDStr)
		http.Error(w, "invalid lotId", http.StatusBadRequest)
		return
	}

	lotID := uint64(lotID64)

	sess, err := s.server.Upgrade(w, r)
	if err != nil {
		log.Println("WebTransport upgrade failed:", err)
		return
	}

	go func(sess *webtransport.Session) {
		ws := &wtSession{session: sess}

		client := &Client{
			LotID:   lotID,
			Session: ws,
		}

		s.hub.Add(client)

		defer func() {
			s.hub.Remove(ws)
			_ = ws.Close()
		}()

		// Đọc incoming streams từ client (chủ yếu để detect disconnect).
		// Client chỉ gửi dữ liệu tối thiểu, server discard tất cả.
		for {
			stream, err := sess.AcceptStream(context.Background())
			if err != nil {
				log.Println("WebTransport accept stream stopped:", err)
				return
			}

			_, _ = io.Copy(io.Discard, stream)
			_ = stream.Close()
		}
	}(sess)
}
