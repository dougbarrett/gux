//go:build js && wasm

package ui

import (
	"syscall/js"

	"github.com/dougbarrett/gux/core"
)

const (
	defaultVideoJSCSS    = "https://vjs.zencdn.net/8.6.1/video-js.css"
	defaultVideoJSScript = "https://vjs.zencdn.net/8.6.1/video.min.js"
	googleIMASDK         = "https://imasdk.googleapis.com/js/sdkloader/ima3.js"
)

// createVideoJSDisposer returns an OnUnmount callback that disposes the VideoJS player.
func createVideoJSDisposer(videoID string) func() {
	return func() {
		videojs := js.Global().Get("videojs")
		if videojs.IsUndefined() || videojs.IsNull() {
			return
		}
		players := videojs.Get("players")
		if players.IsUndefined() || players.IsNull() {
			return
		}
		existing := players.Get(videoID)
		if existing.IsUndefined() || existing.IsNull() {
			return
		}
		existing.Call("dispose")
	}
}

// createVideoJSInitializer returns an OnMount callback that loads and initializes VideoJS.
func createVideoJSInitializer(videoID string, props VideoPlayerProps) func(el any) {
	return func(el any) {
		jsEl := el.(js.Value)

		// Determine CSS and JS URLs
		cssURL := defaultVideoJSCSS
		if props.VideoJSCSS != "" {
			cssURL = props.VideoJSCSS
		}
		jsURL := defaultVideoJSScript
		if props.VideoJSScript != "" {
			jsURL = props.VideoJSScript
		}

		// Load CSS first, then JS
		core.LoadCSS(cssURL, func(err error) {
			if err != nil {
				println("[gux VideoPlayer] Failed to load VideoJS CSS:", err.Error())
				return
			}

			core.LoadScript(jsURL, func(err error) {
				if err != nil {
					println("[gux VideoPlayer] Failed to load VideoJS:", err.Error())
					return
				}

				// Load IMA SDK if ads enabled
				if props.EnableAds && props.AdTagURL != "" {
					core.LoadScript(googleIMASDK, func(err error) {
						if err != nil {
							println("[gux VideoPlayer] Failed to load IMA SDK:", err.Error())
						}
						// Initialize player (with or without ads)
						initVideoJS(videoID, jsEl, props)
					})
				} else {
					initVideoJS(videoID, jsEl, props)
				}
			})
		})
	}
}

// initVideoJS initializes the VideoJS player on the given element.
func initVideoJS(videoID string, el js.Value, props VideoPlayerProps) {
	videojs := js.Global().Get("videojs")
	if videojs.IsUndefined() || videojs.IsNull() {
		println("[gux VideoPlayer] videojs global not available")
		return
	}

	// Build options object
	options := map[string]any{
		"controls": props.Controls,
		"autoplay": props.AutoPlay,
		"preload":  props.Preload,
		"loop":     props.Loop,
		"muted":    props.Muted,
	}
	if props.Poster != "" {
		options["poster"] = props.Poster
	}

	jsOpts := js.Global().Get("JSON").Call("parse", toJSON(options))
	player := videojs.Invoke(videoID, jsOpts)

	// Attach event handlers
	attachPlayerEvents(player, props)

	// Initialize Google IMA ads if enabled
	if props.EnableAds && props.AdTagURL != "" {
		initGoogleIMA(videoID, el, player, props)
	}

	// Fire OnReady
	if props.OnReady != nil {
		player.Call("ready", js.FuncOf(func(this js.Value, args []js.Value) any {
			props.OnReady()
			return nil
		}))
	}
}

// attachPlayerEvents wires VideoJS player events to Go callbacks.
func attachPlayerEvents(player js.Value, props VideoPlayerProps) {
	if props.OnPlay != nil {
		player.Call("on", "play", js.FuncOf(func(this js.Value, args []js.Value) any {
			props.OnPlay()
			return nil
		}))
	}
	if props.OnPause != nil {
		player.Call("on", "pause", js.FuncOf(func(this js.Value, args []js.Value) any {
			props.OnPause()
			return nil
		}))
	}
	if props.OnEnded != nil {
		player.Call("on", "ended", js.FuncOf(func(this js.Value, args []js.Value) any {
			props.OnEnded()
			return nil
		}))
	}
	if props.OnTimeUpdate != nil {
		player.Call("on", "timeupdate", js.FuncOf(func(this js.Value, args []js.Value) any {
			ct := player.Call("currentTime").Float()
			props.OnTimeUpdate(ct)
			return nil
		}))
	}
	if props.OnError != nil {
		player.Call("on", "error", js.FuncOf(func(this js.Value, args []js.Value) any {
			errObj := player.Call("error")
			msg := "Unknown video error"
			if !errObj.IsNull() && !errObj.IsUndefined() {
				m := errObj.Get("message")
				if !m.IsUndefined() {
					msg = m.String()
				}
			}
			props.OnError(msg)
			return nil
		}))
	}
}

// extractAdInfo extracts ad metadata from an IMA ad object into an AdInfo struct.
func extractAdInfo(ad js.Value) AdInfo {
	info := AdInfo{}
	if ad.IsNull() || ad.IsUndefined() {
		return info
	}
	if id := ad.Call("getAdId"); !id.IsUndefined() {
		info.AdID = id.String()
	}
	if title := ad.Call("getTitle"); !title.IsUndefined() {
		info.AdTitle = title.String()
	}
	if sys := ad.Call("getAdSystem"); !sys.IsUndefined() {
		info.AdSystem = sys.String()
	}
	info.Duration = ad.Call("getDuration").Float()
	info.IsSkippable = ad.Call("isSkippable").Bool()
	if ct := ad.Call("getContentType"); !ct.IsUndefined() {
		info.ContentType = ct.String()
	}
	info.AdPosition = ad.Call("getAdPodInfo").Call("getAdPosition").Int()
	info.TotalAds = ad.Call("getAdPodInfo").Call("getTotalAds").Int()
	return info
}

// initGoogleIMA sets up Google IMA ads on the video player.
func initGoogleIMA(videoID string, videoEl js.Value, player js.Value, props VideoPlayerProps) {
	google := js.Global().Get("google")
	if google.IsUndefined() || google.IsNull() {
		println("[gux VideoPlayer] Google IMA SDK not available")
		return
	}
	ima := google.Get("ima")
	if ima.IsUndefined() || ima.IsNull() {
		println("[gux VideoPlayer] google.ima not available")
		return
	}

	doc := js.Global().Get("document")

	// Create ad container overlay
	adContainer := doc.Call("createElement", "div")
	adContainer.Set("id", videoID+"-ad-container")
	style := adContainer.Get("style")
	style.Set("position", "absolute")
	style.Set("top", "0")
	style.Set("left", "0")
	style.Set("width", "100%")
	style.Set("height", "100%")
	style.Set("pointerEvents", "none")

	// Insert ad container into player element's parent
	parent := videoEl.Get("parentNode")
	if parent.IsNull() || parent.IsUndefined() {
		println("[gux VideoPlayer] Video element has no parent, cannot attach ad container")
		return
	}
	parentStyle := parent.Get("style")
	parentStyle.Set("position", "relative")
	parent.Call("appendChild", adContainer)

	// Initialize IMA
	adDisplayContainer := ima.Get("AdDisplayContainer").New(adContainer, videoEl)
	adDisplayContainer.Call("initialize")

	adsLoader := ima.Get("AdsLoader").New(adDisplayContainer)

	// Handle ads manager loaded
	adsLoader.Call("addEventListener",
		ima.Get("AdsManagerLoadedEvent").Get("Type").Get("ADS_MANAGER_LOADED"),
		js.FuncOf(func(this js.Value, args []js.Value) any {
			event := args[0]
			adsManager := event.Call("getAdsManager", videoEl)

			// Wire up IMA ad events to Go callbacks
			adEventType := ima.Get("AdEvent").Get("Type")

			if props.OnAdLoaded != nil {
				adsManager.Call("addEventListener",
					adEventType.Get("LOADED"),
					js.FuncOf(func(this js.Value, args []js.Value) any {
						ad := args[0].Call("getAd")
						props.OnAdLoaded(extractAdInfo(ad))
						return nil
					}), false)
			}

			if props.OnAdStarted != nil {
				adsManager.Call("addEventListener",
					adEventType.Get("STARTED"),
					js.FuncOf(func(this js.Value, args []js.Value) any {
						ad := args[0].Call("getAd")
						props.OnAdStarted(extractAdInfo(ad))
						return nil
					}), false)
			}

			if props.OnAdComplete != nil {
				adsManager.Call("addEventListener",
					adEventType.Get("COMPLETE"),
					js.FuncOf(func(this js.Value, args []js.Value) any {
						ad := args[0].Call("getAd")
						props.OnAdComplete(extractAdInfo(ad))
						return nil
					}), false)
			}

			if props.OnAdSkipped != nil {
				adsManager.Call("addEventListener",
					adEventType.Get("SKIPPED"),
					js.FuncOf(func(this js.Value, args []js.Value) any {
						ad := args[0].Call("getAd")
						props.OnAdSkipped(extractAdInfo(ad))
						return nil
					}), false)
			}

			if props.OnAdProgress != nil {
				adsManager.Call("addEventListener",
					adEventType.Get("AD_PROGRESS"),
					js.FuncOf(func(this js.Value, args []js.Value) any {
						ad := args[0].Call("getAd")
						info := extractAdInfo(ad)
						// Get current time from the ad data
						adData := args[0].Call("getAdData")
						if !adData.IsUndefined() && !adData.IsNull() {
							ct := adData.Get("currentTime")
							if !ct.IsUndefined() {
								info.CurrentTime = ct.Float()
							}
						}
						props.OnAdProgress(info)
						return nil
					}), false)
			}

			if props.OnAdClicked != nil {
				adsManager.Call("addEventListener",
					adEventType.Get("CLICK"),
					js.FuncOf(func(this js.Value, args []js.Value) any {
						ad := args[0].Call("getAd")
						info := extractAdInfo(ad)
						// Try to get click-through URL
						survey := ad.Call("getSurveyUrl")
						if !survey.IsUndefined() && survey.String() != "" {
							info.ClickURL = survey.String()
						}
						props.OnAdClicked(info)
						return nil
					}), false)
			}

			if props.OnContentPause != nil {
				adsManager.Call("addEventListener",
					adEventType.Get("CONTENT_PAUSE_REQUESTED"),
					js.FuncOf(func(this js.Value, args []js.Value) any {
						props.OnContentPause()
						return nil
					}), false)
			}

			if props.OnContentResume != nil {
				adsManager.Call("addEventListener",
					adEventType.Get("CONTENT_RESUME_REQUESTED"),
					js.FuncOf(func(this js.Value, args []js.Value) any {
						props.OnContentResume()
						return nil
					}), false)
			}

			// Start ads when content plays
			player.Call("on", "play", js.FuncOf(func(this js.Value, args []js.Value) any {
				adsManager.Call("init",
					videoEl.Get("offsetWidth"),
					videoEl.Get("offsetHeight"),
					ima.Get("ViewMode").Get("NORMAL"),
				)
				adsManager.Call("start")
				return nil
			}))

			return nil
		}),
		false,
	)

	// Handle ads error — fire callback then play content
	adsLoader.Call("addEventListener",
		ima.Get("AdErrorEvent").Get("Type").Get("AD_ERROR"),
		js.FuncOf(func(this js.Value, args []js.Value) any {
			errMsg := "Unknown ad error"
			if len(args) > 0 {
				errObj := args[0].Call("getError")
				if !errObj.IsUndefined() && !errObj.IsNull() {
					msg := errObj.Call("getMessage")
					if !msg.IsUndefined() {
						errMsg = msg.String()
					}
				}
			}
			println("[gux VideoPlayer] IMA ad error:", errMsg)
			if props.OnAdError != nil {
				props.OnAdError(errMsg)
			}
			player.Call("play")
			return nil
		}),
		false,
	)

	// Request ads
	adsRequest := ima.Get("AdsRequest").New()
	adsRequest.Set("adTagUrl", props.AdTagURL)
	adsRequest.Set("linearAdSlotWidth", videoEl.Get("offsetWidth"))
	adsRequest.Set("linearAdSlotHeight", videoEl.Get("offsetHeight"))

	adsLoader.Call("requestAds", adsRequest)
}

// createNativeVideoHandlers returns an OnMount callback for HTML5 video events (no VideoJS).
func createNativeVideoHandlers(props VideoPlayerProps) func(el any) {
	return func(el any) {
		v := el.(js.Value)

		if props.OnPlay != nil {
			v.Call("addEventListener", "play", js.FuncOf(func(this js.Value, args []js.Value) any {
				props.OnPlay()
				return nil
			}))
		}
		if props.OnPause != nil {
			v.Call("addEventListener", "pause", js.FuncOf(func(this js.Value, args []js.Value) any {
				props.OnPause()
				return nil
			}))
		}
		if props.OnEnded != nil {
			v.Call("addEventListener", "ended", js.FuncOf(func(this js.Value, args []js.Value) any {
				props.OnEnded()
				return nil
			}))
		}
		if props.OnTimeUpdate != nil {
			v.Call("addEventListener", "timeupdate", js.FuncOf(func(this js.Value, args []js.Value) any {
				ct := v.Get("currentTime").Float()
				props.OnTimeUpdate(ct)
				return nil
			}))
		}
		if props.OnError != nil {
			v.Call("addEventListener", "error", js.FuncOf(func(this js.Value, args []js.Value) any {
				errObj := v.Get("error")
				msg := "Unknown video error"
				if !errObj.IsNull() && !errObj.IsUndefined() {
					code := errObj.Get("code").Int()
					switch code {
					case 1:
						msg = "MEDIA_ERR_ABORTED: Fetching was aborted"
					case 2:
						msg = "MEDIA_ERR_NETWORK: Network error"
					case 3:
						msg = "MEDIA_ERR_DECODE: Decoding error"
					case 4:
						msg = "MEDIA_ERR_SRC_NOT_SUPPORTED: Source not supported"
					}
				}
				props.OnError(msg)
				return nil
			}))
		}
		if props.OnReady != nil {
			v.Call("addEventListener", "loadedmetadata", js.FuncOf(func(this js.Value, args []js.Value) any {
				props.OnReady()
				return nil
			}))
		}
	}
}

// toJSON converts a Go map to a JSON string for passing to JS.
func toJSON(m map[string]any) string {
	// Simple manual JSON for known types
	result := "{"
	first := true
	for k, v := range m {
		if !first {
			result += ","
		}
		first = false
		result += `"` + k + `":`
		switch val := v.(type) {
		case bool:
			if val {
				result += "true"
			} else {
				result += "false"
			}
		case string:
			result += `"` + val + `"`
		case int:
			result += js.Global().Get("String").Invoke(val).String()
		default:
			result += "null"
		}
	}
	result += "}"
	return result
}
