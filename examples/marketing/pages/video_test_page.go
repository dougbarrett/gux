package pages

import (
	"fmt"

	"github.com/dougbarrett/gux/core"
	"github.com/dougbarrett/gux/ui"
)

// VideoDemo is a test page for the VideoPlayer component.
func VideoDemo(r *core.Router) func() core.Node {
	adLog := r.StateString("adLog", "Waiting for ad events...")

	return func() core.Node {
		return MarketingLayout(r,
			core.Section(core.Class("py-16 px-4 max-w-4xl mx-auto"),
				core.H1(core.Class("text-3xl font-bold text-gray-900 dark:text-white mb-8"),
					core.Text("Video Player Test"),
				),

				// Test 1: Basic HTML5 video with controls
				core.Div(core.Attrs{ID: "test-native", Class: "mb-12"},
					core.H2(core.Class("text-xl font-semibold text-gray-700 dark:text-gray-300 mb-4"),
						core.Text("1. Native HTML5 Video"),
					),
					core.Video(core.Attrs{
						ID:       "native-video",
						Controls: true,
						Muted:    true,
						Preload:  "metadata",
						Width:    "640",
						Height:   "360",
					},
						core.Source(core.Attrs{
							Src:  "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4",
							Type: "video/mp4",
						}),
						core.P(core.Class("text-gray-500"),
							core.Text("Your browser does not support the video tag."),
						),
					),
				),

				// Test 2: VideoPlayer component (no VideoJS)
				core.Div(core.Attrs{ID: "test-component", Class: "mb-12"},
					core.H2(core.Class("text-xl font-semibold text-gray-700 dark:text-gray-300 mb-4"),
						core.Text("2. VideoPlayer Component (Native)"),
					),
					ui.VideoPlayer(ui.VideoPlayerProps{
						Sources: []ui.VideoSource{
							{Src: "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ElephantsDream.mp4", Type: "video/mp4"},
						},
						Controls: true,
						Muted:    true,
						Width:    "640",
						Height:   "360",
					}),
				),

				// Test 3: VideoPlayer with VideoJS enabled
				core.Div(core.Attrs{ID: "test-videojs", Class: "mb-12"},
					core.H2(core.Class("text-xl font-semibold text-gray-700 dark:text-gray-300 mb-4"),
						core.Text("3. VideoPlayer with VideoJS"),
					),
					ui.VideoPlayer(ui.VideoPlayerProps{
						Sources: []ui.VideoSource{
							{Src: "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerBlazes.mp4", Type: "video/mp4"},
						},
						EnableVideoJS: true,
						Controls:      true,
						Muted:         true,
						Width:         "640",
						Height:        "360",
						OnPlay: func() {
							println("[Test 3] Video started playing")
						},
						OnPause: func() {
							println("[Test 3] Video paused")
						},
					}),
				),

				// Test 4: VideoPlayer with VideoJS + Google IMA ads + event handlers
				core.Div(core.Attrs{ID: "test-ima", Class: "mb-12"},
					core.H2(core.Class("text-xl font-semibold text-gray-700 dark:text-gray-300 mb-4"),
						core.Text("4. VideoPlayer with Google IMA Ads"),
					),
					ui.VideoPlayer(ui.VideoPlayerProps{
						Sources: []ui.VideoSource{
							{Src: "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/Sintel.mp4", Type: "video/mp4"},
						},
						EnableVideoJS: true,
						EnableAds:     true,
						AdTagURL:      "https://pubads.g.doubleclick.net/gampad/ads?iu=/21775744923/external/single_ad_samples&sz=640x480&cust_params=sample_ct%3Dlinear&ciu_szs=300x250%2C728x90&gdfp_req=1&output=vast&unviewed_position_start=1&env=vp&impl=s&correlator=",
						Controls:      true,
						Muted:         true,
						Width:         "640",
						Height:        "360",

						// Video events
						OnReady: func() {
							adLog.Set("Player ready")
						},
						OnPlay: func() {
							println("[Test 4] Content playing")
						},

						// IMA ad events
						OnAdLoaded: func(info ui.AdInfo) {
							adLog.Set(fmt.Sprintf("Ad loaded: %q (%s) %.0fs", info.AdTitle, info.AdSystem, info.Duration))
						},
						OnAdStarted: func(info ui.AdInfo) {
							adLog.Set(fmt.Sprintf("Ad started: %q (%.0fs, skippable: %v)", info.AdTitle, info.Duration, info.IsSkippable))
						},
						OnAdComplete: func(info ui.AdInfo) {
							adLog.Set(fmt.Sprintf("Ad complete: %q", info.AdTitle))
						},
						OnAdSkipped: func(info ui.AdInfo) {
							adLog.Set(fmt.Sprintf("Ad skipped: %q", info.AdTitle))
						},
						OnAdError: func(errMsg string) {
							adLog.Set(fmt.Sprintf("Ad error: %s", errMsg))
						},
						OnContentPause: func() {
							println("[Test 4] Content paused for ad break")
						},
						OnContentResume: func() {
							adLog.Set("Content resumed after ad break")
						},
					}),

					// Ad event log display
					core.Div(core.Attrs{ID: "ad-event-log", Class: "mt-4 p-4 bg-gray-100 dark:bg-gray-800 rounded-lg"},
						core.H3(core.Class("text-sm font-semibold text-gray-600 dark:text-gray-400 mb-2"),
							core.Text("Ad Event Log:"),
						),
						core.P(core.Attrs{ID: "ad-log-text", Class: "text-sm font-mono text-gray-800 dark:text-gray-200"},
							core.Text(adLog.Get()),
						),
					),
				),
			),
		)
	}
}
