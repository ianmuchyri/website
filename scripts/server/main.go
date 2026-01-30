package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gorilla/websocket"
)

const reloaderScript = `
<script>
(function() {
	const ws = new WebSocket('ws://' + window.location.host + '/ws');
	ws.onmessage = function(event) {
		if (event.data === 'reload') {
			console.log('Files changed, reloading...');
			window.location.reload();
		}
	};
	ws.onclose = function() {
		console.log('Dev server disconnected, retrying...');
		setTimeout(() => window.location.reload(), 1000);
	};
})();
</script>`

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

// Inject live reload script into HTML responses
type liveReloadInjector struct {
	http.ResponseWriter
	injected bool
}

func (w *liveReloadInjector) WriteHeader(statusCode int) {
	if strings.Contains(w.Header().Get("Content-Type"), "text/html") {
		w.Header().Del("Content-Length")
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *liveReloadInjector) Write(b []byte) (int, error) {
	if !w.injected && strings.Contains(w.Header().Get("Content-Type"), "text/html") {
		w.Header().Del("Content-Length") // Ensure it's deleted if WriteHeader wasn't called explicitly

		// Inject the reload script before </body>
		content := string(b)

		if idx := strings.LastIndex(content, "</body>"); idx != -1 {
			content = content[:idx] + reloaderScript + content[idx:]
			w.injected = true
			return w.ResponseWriter.Write([]byte(content))
		}
	}
	return w.ResponseWriter.Write(b)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	log.Printf("Browser connected (total: %d)", len(clients))

	// Keep connection alive
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}

	clientsMu.Lock()
	delete(clients, conn)
	clientsMu.Unlock()
	log.Printf("Browser disconnected (total: %d)", len(clients))
}

func notifyClients() {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for conn := range clients {
		if err := conn.WriteMessage(websocket.TextMessage, []byte("reload")); err != nil {
			log.Printf("Error notifying client: %v", err)
			conn.Close()
			delete(clients, conn)
		}
	}
	log.Printf("Notified %d browser(s) to reload", len(clients))
}

func watchFiles(watchDir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Recursively watch the directory
	err = filepath.Walk(watchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// Skip .git and node_modules
			if strings.Contains(path, ".git") ||
				strings.Contains(path, "node_modules") ||
				strings.Contains(path, "vendor") {
				return filepath.SkipDir
			}
			if err := watcher.Add(path); err != nil {
				log.Printf("Warning: couldn't watch %s: %v", path, err)
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("Warning: error walking directory: %v", err)
	}

	log.Printf("Watching for changes in: %s", watchDir)

	// Debounce timer to avoid multiple reloads on rapid changes
	var timer *time.Timer

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// Ignore temporary files
			if strings.HasPrefix(filepath.Base(event.Name), ".") ||
				strings.HasSuffix(event.Name, "~") ||
				strings.HasSuffix(event.Name, ".swp") {
				continue
			}

			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				log.Printf("File changed: %s", event.Name)
				// Debounce: wait 200ms before notifying
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(200*time.Millisecond, notifyClients)
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

func main() {
	p := flag.String("port", "8080", "port to run the file server on")
	watch := flag.Bool("watch", true, "enable file watching and auto-reload")
	dir := flag.String("dir", "", "directory to serve files from (default: smart detection)")
	flag.Parse()

	port := *p
	serveDir := *dir

	if serveDir == "" {
		serveDir = "."
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(serveDir)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("File server started at port %s", port)
	log.Printf("Open your browser at http://localhost:%s", port)
	log.Printf("Serving from: %s", absPath)

	// Start file watcher if enabled
	if *watch {
		log.Println("Live reload enabled")
		go watchFiles(absPath)
	}

	// WebSocket endpoint for live reload
	http.HandleFunc("/ws", handleWebSocket)

	// File server with live reload injection
	fileServer := http.FileServer(http.Dir(serveDir))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if *watch {
			// Wrap response writer to inject reload script
			injector := &liveReloadInjector{ResponseWriter: w}
			fileServer.ServeHTTP(injector, r)
		} else {
			fileServer.ServeHTTP(w, r)
		}
	})

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
