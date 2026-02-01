package ui

import (
	"fmt"
	"sync/atomic"

	"github.com/dougbarrett/gux/core"
)

// VideoSource represents a video source with format.
type VideoSource struct {
	Src  string // Video file URL
	Type string // MIME type (e.g., "video/mp4", "video/webm", "video/ogg")
}

// AdInfo contains metadata about a Google IMA ad event.
type AdInfo struct {
	AdID         string  // Unique ad identifier
	AdTitle      string  // Ad creative title
	AdSystem     string  // Ad system (e.g., "GDFP")
	Duration     float64 // Ad duration in seconds
	CurrentTime  float64 // Current playback position in the ad (seconds)
	IsSkippable  bool    // Whether the ad can be skipped
	ContentType  string  // MIME type of the ad media
	AdPosition   int     // Position in the ad pod (1-based)
	TotalAds     int     // Total ads in the pod
	ClickURL     string  // Ad click-through URL
}

// VideoPlayerProps configures a VideoPlayer component.
type VideoPlayerProps struct {
	// Basic video attributes
	Sources  []VideoSource // Video sources (multiple formats for browser compatibility)
	Poster   string        // Poster image URL
	Width    string        // CSS width (default: "100%")
	Height   string        // CSS height (default: "auto")
	AutoPlay bool          // Auto-play video
	Loop     bool          // Loop playback
	Muted    bool          // Start muted
	Controls bool          // Show playback controls (default: true when unset)
	Preload  string        // Preload strategy: "none", "metadata", "auto" (default: "metadata")
	Class    string        // Additional CSS classes

	// VideoJS integration (loads library from CDN on first use)
	EnableVideoJS bool   // Use VideoJS enhanced player
	VideoJSCSS    string // Custom VideoJS CSS URL (default: CDN)
	VideoJSScript string // Custom VideoJS JS URL (default: CDN)

	// Google IMA ads integration (requires EnableVideoJS)
	EnableAds bool   // Enable Google IMA pre-roll/mid-roll ads
	AdTagURL  string // VAST/VMAP ad tag URL

	// Event handlers (WASM only, ignored in SSR)
	OnPlay       func()
	OnPause      func()
	OnEnded      func()
	OnTimeUpdate func(currentTime float64) // Fires periodically during playback (~4x/sec) with current position in seconds
	OnError      func(msg string)
	OnReady      func() // Called when player is fully initialized

	// Google IMA ad event handlers (WASM only, ignored in SSR)
	// These fire during the ad lifecycle when EnableAds is true.
	OnAdLoaded    func(adInfo AdInfo)  // Ad creative loaded and ready
	OnAdStarted   func(adInfo AdInfo)  // Ad playback started
	OnAdComplete  func(adInfo AdInfo)  // Ad playback finished
	OnAdSkipped   func(adInfo AdInfo)  // User skipped the ad
	OnAdError     func(errMsg string)  // Ad failed to load or play
	OnAdProgress  func(adInfo AdInfo)  // Periodic progress during ad playback
	OnAdClicked   func(adInfo AdInfo)  // User clicked the ad
	OnContentResume func()             // Content should resume after ad break
	OnContentPause  func()             // Content should pause for ad break
}

var videoPlayerIDCounter atomic.Int64

// VideoPlayer creates a video player component.
//
// Basic HTML5 video:
//
//	ui.VideoPlayer(ui.VideoPlayerProps{
//	    Sources:  []ui.VideoSource{{Src: "/video.mp4", Type: "video/mp4"}},
//	    Poster:   "/poster.jpg",
//	    Controls: true,
//	})
//
// With VideoJS (enhanced player UI):
//
//	ui.VideoPlayer(ui.VideoPlayerProps{
//	    Sources:       []ui.VideoSource{{Src: "/video.mp4", Type: "video/mp4"}},
//	    EnableVideoJS: true,
//	    Controls:      true,
//	})
//
// With Google IMA ads:
//
//	ui.VideoPlayer(ui.VideoPlayerProps{
//	    Sources:       []ui.VideoSource{{Src: "/video.mp4", Type: "video/mp4"}},
//	    EnableVideoJS: true,
//	    EnableAds:     true,
//	    AdTagURL:      "https://pubads.g.doubleclick.net/gampad/ads?...",
//	    Controls:      true,
//	})
func VideoPlayer(props VideoPlayerProps) core.Node {
	if props.Preload == "" {
		props.Preload = "metadata"
	}
	if props.Width == "" {
		props.Width = "100%"
	}
	if props.Height == "" {
		props.Height = "auto"
	}

	videoID := fmt.Sprintf("gux-video-%d", videoPlayerIDCounter.Add(1))

	cssClass := "gux-video-player"
	if props.EnableVideoJS {
		cssClass = MergeClasses(cssClass, "video-js", "vjs-default-skin")
	}
	cssClass = MergeClasses(cssClass, props.Class)

	attrs := core.Attrs{
		ID:       videoID,
		Class:    cssClass,
		Poster:   props.Poster,
		Width:    props.Width,
		Height:   props.Height,
		Autoplay: props.AutoPlay,
		Loop:     props.Loop,
		Muted:    props.Muted,
		Controls: props.Controls,
		Preload:  props.Preload,
	}

	// Set OnMount for WASM initialization (VideoJS or native events)
	if props.EnableVideoJS {
		attrs.OnMount = createVideoJSInitializer(videoID, props)
	} else if hasEventHandlers(props) {
		attrs.OnMount = createNativeVideoHandlers(props)
	}

	// Build source children
	var children []core.Node
	for _, src := range props.Sources {
		children = append(children, core.Source(core.Attrs{
			Src:  src.Src,
			Type: src.Type,
		}))
	}

	// Fallback text
	children = append(children,
		core.P(core.Class("vjs-no-js"),
			core.Text("Your browser does not support the video tag."),
		),
	)

	return core.Video(attrs, children...)
}

func hasEventHandlers(props VideoPlayerProps) bool {
	return props.OnPlay != nil || props.OnPause != nil ||
		props.OnEnded != nil || props.OnTimeUpdate != nil ||
		props.OnError != nil || props.OnReady != nil ||
		props.OnAdLoaded != nil || props.OnAdStarted != nil || props.OnAdComplete != nil
}
