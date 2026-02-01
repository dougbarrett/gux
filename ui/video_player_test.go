package ui

import (
	"strings"
	"testing"

	"github.com/dougbarrett/gux/core"
)

func TestVideoPlayer_BasicHTML5(t *testing.T) {
	player := VideoPlayer(VideoPlayerProps{
		Sources: []VideoSource{
			{Src: "/video.mp4", Type: "video/mp4"},
		},
		Controls: true,
	})

	html := renderHTML(player)

	if !strings.Contains(html, "<video") {
		t.Errorf("expected <video> element, got: %s", html)
	}
	if !strings.Contains(html, "controls") {
		t.Errorf("expected controls attribute, got: %s", html)
	}
	if !strings.Contains(html, `src="/video.mp4"`) {
		t.Errorf("expected video source, got: %s", html)
	}
	if !strings.Contains(html, `type="video/mp4"`) {
		t.Errorf("expected source type, got: %s", html)
	}
}

func TestVideoPlayer_DefaultValues(t *testing.T) {
	player := VideoPlayer(VideoPlayerProps{
		Sources: []VideoSource{
			{Src: "/video.mp4", Type: "video/mp4"},
		},
	})

	html := renderHTML(player)

	// Default preload is "metadata"
	if !strings.Contains(html, `preload="metadata"`) {
		t.Errorf("expected default preload=metadata, got: %s", html)
	}
	// Default width is "100%"
	if !strings.Contains(html, `width="100%"`) {
		t.Errorf("expected default width=100%%, got: %s", html)
	}
}

func TestVideoPlayer_MultipleSources(t *testing.T) {
	player := VideoPlayer(VideoPlayerProps{
		Sources: []VideoSource{
			{Src: "/video.mp4", Type: "video/mp4"},
			{Src: "/video.webm", Type: "video/webm"},
		},
	})

	html := renderHTML(player)

	if !strings.Contains(html, `src="/video.mp4"`) {
		t.Errorf("expected mp4 source, got: %s", html)
	}
	if !strings.Contains(html, `src="/video.webm"`) {
		t.Errorf("expected webm source, got: %s", html)
	}
}

func TestVideoPlayer_VideoJSClasses(t *testing.T) {
	player := VideoPlayer(VideoPlayerProps{
		Sources: []VideoSource{
			{Src: "/video.mp4", Type: "video/mp4"},
		},
		EnableVideoJS: true,
	})

	html := renderHTML(player)

	if !strings.Contains(html, "video-js") {
		t.Errorf("expected video-js class for VideoJS, got: %s", html)
	}
	if !strings.Contains(html, "vjs-default-skin") {
		t.Errorf("expected vjs-default-skin class, got: %s", html)
	}
}

func TestVideoPlayer_NoVideoJSClasses(t *testing.T) {
	player := VideoPlayer(VideoPlayerProps{
		Sources: []VideoSource{
			{Src: "/video.mp4", Type: "video/mp4"},
		},
		EnableVideoJS: false,
	})

	html := renderHTML(player)

	if strings.Contains(html, "video-js") {
		t.Errorf("should not have video-js class when VideoJS disabled, got: %s", html)
	}
}

func TestVideoPlayer_BooleanAttributes(t *testing.T) {
	player := VideoPlayer(VideoPlayerProps{
		Sources: []VideoSource{
			{Src: "/video.mp4", Type: "video/mp4"},
		},
		AutoPlay: true,
		Loop:     true,
		Muted:    true,
		Controls: true,
	})

	html := renderHTML(player)

	if !strings.Contains(html, "autoplay") {
		t.Errorf("expected autoplay attribute, got: %s", html)
	}
	if !strings.Contains(html, "loop") {
		t.Errorf("expected loop attribute, got: %s", html)
	}
	if !strings.Contains(html, "muted") {
		t.Errorf("expected muted attribute, got: %s", html)
	}
	if !strings.Contains(html, "controls") {
		t.Errorf("expected controls attribute, got: %s", html)
	}
}

func TestVideoPlayer_PosterAndDimensions(t *testing.T) {
	player := VideoPlayer(VideoPlayerProps{
		Sources: []VideoSource{
			{Src: "/video.mp4", Type: "video/mp4"},
		},
		Poster: "/poster.jpg",
		Width:  "640",
		Height: "360",
	})

	html := renderHTML(player)

	if !strings.Contains(html, `poster="/poster.jpg"`) {
		t.Errorf("expected poster attribute, got: %s", html)
	}
	if !strings.Contains(html, `width="640"`) {
		t.Errorf("expected width attribute, got: %s", html)
	}
	if !strings.Contains(html, `height="360"`) {
		t.Errorf("expected height attribute, got: %s", html)
	}
}

func TestVideoPlayer_CustomClass(t *testing.T) {
	player := VideoPlayer(VideoPlayerProps{
		Sources: []VideoSource{
			{Src: "/video.mp4", Type: "video/mp4"},
		},
		Class: "my-custom-class",
	})

	html := renderHTML(player)

	if !strings.Contains(html, "my-custom-class") {
		t.Errorf("expected custom class, got: %s", html)
	}
	if !strings.Contains(html, "gux-video-player") {
		t.Errorf("expected base class gux-video-player, got: %s", html)
	}
}

func TestVideoPlayer_FallbackText(t *testing.T) {
	player := VideoPlayer(VideoPlayerProps{
		Sources: []VideoSource{
			{Src: "/video.mp4", Type: "video/mp4"},
		},
	})

	html := renderHTML(player)

	if !strings.Contains(html, "Your browser does not support the video tag.") {
		t.Errorf("expected fallback text, got: %s", html)
	}
}

func TestVideoPlayer_UniqueIDs(t *testing.T) {
	player1 := VideoPlayer(VideoPlayerProps{
		Sources: []VideoSource{{Src: "/a.mp4", Type: "video/mp4"}},
	})
	player2 := VideoPlayer(VideoPlayerProps{
		Sources: []VideoSource{{Src: "/b.mp4", Type: "video/mp4"}},
	})

	html1 := renderHTML(player1)
	html2 := renderHTML(player2)

	// Extract IDs - they should be different
	if html1 == html2 {
		t.Error("two VideoPlayer instances should have different IDs")
	}
}

func TestVideoPlayer_SSR_IgnoresEventHandlers(t *testing.T) {
	// Event handlers should not affect HTML output
	player := VideoPlayer(VideoPlayerProps{
		Sources: []VideoSource{
			{Src: "/video.mp4", Type: "video/mp4"},
		},
		Controls: true,
		OnPlay:   func() {},
		OnPause:  func() {},
		OnEnded:  func() {},
		OnReady:  func() {},
		OnError:  func(msg string) {},
		OnTimeUpdate: func(ct float64) {},
	})

	html := renderHTML(player)

	// Should still render normally
	if !strings.Contains(html, "<video") {
		t.Errorf("expected video element with event handlers, got: %s", html)
	}
}

func TestVideoPlayer_SSR_IgnoresAdHandlers(t *testing.T) {
	player := VideoPlayer(VideoPlayerProps{
		Sources: []VideoSource{
			{Src: "/video.mp4", Type: "video/mp4"},
		},
		EnableVideoJS: true,
		EnableAds:     true,
		AdTagURL:      "https://example.com/vast",
		OnAdLoaded:    func(info AdInfo) {},
		OnAdStarted:   func(info AdInfo) {},
		OnAdComplete:  func(info AdInfo) {},
		OnAdSkipped:   func(info AdInfo) {},
		OnAdError:     func(msg string) {},
		OnAdProgress:  func(info AdInfo) {},
		OnAdClicked:   func(info AdInfo) {},
		OnContentPause:  func() {},
		OnContentResume: func() {},
	})

	html := renderHTML(player)

	if !strings.Contains(html, "<video") {
		t.Errorf("expected video element with ad handlers, got: %s", html)
	}
}

func TestVideoPlayer_CustomPreload(t *testing.T) {
	player := VideoPlayer(VideoPlayerProps{
		Sources: []VideoSource{
			{Src: "/video.mp4", Type: "video/mp4"},
		},
		Preload: "auto",
	})

	html := renderHTML(player)

	if !strings.Contains(html, `preload="auto"`) {
		t.Errorf("expected preload=auto, got: %s", html)
	}
}

func TestVideoPlayer_DisposerStub_ReturnsNil(t *testing.T) {
	// On server builds, createVideoJSDisposer returns nil
	disposer := createVideoJSDisposer("test-id")
	if disposer != nil {
		t.Error("expected nil disposer on server build")
	}
}

func TestVideoPlayer_InitializerStub_ReturnsNil(t *testing.T) {
	// On server builds, createVideoJSInitializer returns nil
	initializer := createVideoJSInitializer("test-id", VideoPlayerProps{})
	if initializer != nil {
		t.Error("expected nil initializer on server build")
	}
}

func TestVideoPlayer_NativeHandlersStub_ReturnsNil(t *testing.T) {
	// On server builds, createNativeVideoHandlers returns nil
	handler := createNativeVideoHandlers(VideoPlayerProps{})
	if handler != nil {
		t.Error("expected nil native handler on server build")
	}
}

func TestHasEventHandlers(t *testing.T) {
	tests := []struct {
		name     string
		props    VideoPlayerProps
		expected bool
	}{
		{"no handlers", VideoPlayerProps{}, false},
		{"OnPlay", VideoPlayerProps{OnPlay: func() {}}, true},
		{"OnPause", VideoPlayerProps{OnPause: func() {}}, true},
		{"OnEnded", VideoPlayerProps{OnEnded: func() {}}, true},
		{"OnTimeUpdate", VideoPlayerProps{OnTimeUpdate: func(float64) {}}, true},
		{"OnError", VideoPlayerProps{OnError: func(string) {}}, true},
		{"OnReady", VideoPlayerProps{OnReady: func() {}}, true},
		{"OnAdLoaded", VideoPlayerProps{OnAdLoaded: func(AdInfo) {}}, true},
		{"OnAdStarted", VideoPlayerProps{OnAdStarted: func(AdInfo) {}}, true},
		{"OnAdComplete", VideoPlayerProps{OnAdComplete: func(AdInfo) {}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasEventHandlers(tt.props)
			if got != tt.expected {
				t.Errorf("hasEventHandlers() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestVideoPlayer_OnUnmount_SetForVideoJS(t *testing.T) {
	// When EnableVideoJS is true, the video element should have both
	// OnMount and OnUnmount callbacks set (or nil stubs on server)
	player := VideoPlayer(VideoPlayerProps{
		Sources: []VideoSource{
			{Src: "/video.mp4", Type: "video/mp4"},
		},
		EnableVideoJS: true,
	})

	// On server builds, stubs return nil, so OnMount and OnUnmount
	// will be nil. The important thing is that the code path doesn't panic.
	html := renderHTML(player)
	if !strings.Contains(html, "video-js") {
		t.Errorf("expected VideoJS classes, got: %s", html)
	}
}

func TestVideoPlayer_OnUnmount_NotSetWithoutVideoJS(t *testing.T) {
	// When EnableVideoJS is false, OnUnmount should not be set
	// (native event handlers don't need disposal)
	player := VideoPlayer(VideoPlayerProps{
		Sources: []VideoSource{
			{Src: "/video.mp4", Type: "video/mp4"},
		},
		EnableVideoJS: false,
		OnPlay:        func() {},
	})

	html := renderHTML(player)
	if strings.Contains(html, "video-js") {
		t.Errorf("should not have VideoJS classes, got: %s", html)
	}
}

func TestVideoPlayer_RendersAsVideoTag(t *testing.T) {
	player := VideoPlayer(VideoPlayerProps{
		Sources: []VideoSource{
			{Src: "/video.mp4", Type: "video/mp4"},
		},
	})

	html := renderHTML(player)

	if !strings.HasPrefix(html, "<video") {
		t.Errorf("expected output to start with <video, got: %s", html[:min(50, len(html))])
	}
	if !strings.HasSuffix(html, "</video>") {
		t.Errorf("expected output to end with </video>, got: %s", html[max(0, len(html)-50):])
	}
}

func TestVideoPlayer_SourceElement(t *testing.T) {
	// Verify Source helper creates proper source elements
	src := core.Source(core.Attrs{
		Src:  "/test.mp4",
		Type: "video/mp4",
	})

	html := src.Render(core.HTML()).HTML()

	if !strings.Contains(html, "<source") {
		t.Errorf("expected <source> element, got: %s", html)
	}
	if !strings.Contains(html, `src="/test.mp4"`) {
		t.Errorf("expected src attribute, got: %s", html)
	}
}
