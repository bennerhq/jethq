package app

import "testing"

func TestInitialWindowSizeForMonitor(t *testing.T) {
	tests := []struct {
		name          string
		monitorWidth  int
		monitorHeight int
		browseMode    bool
		wantWidth     int
		wantHeight    int
	}{
		{
			name:          "unknown monitor uses default connected size",
			monitorWidth:  0,
			monitorHeight: 0,
			wantWidth:     DefaultWindowWidth,
			wantHeight:    DefaultWindowHeight,
		},
		{
			name:          "unknown monitor uses browse size",
			monitorWidth:  0,
			monitorHeight: 0,
			browseMode:    true,
			wantWidth:     BrowseWindowWidth,
			wantHeight:    BrowseWindowHeight,
		},
		{
			name:          "large monitor prefers full hd when connected",
			monitorWidth:  2560,
			monitorHeight: 1440,
			wantWidth:     targetWindowWidth,
			wantHeight:    targetWindowHeight,
		},
		{
			name:          "large monitor keeps shorter browse height",
			monitorWidth:  2560,
			monitorHeight: 1440,
			browseMode:    true,
			wantWidth:     BrowseWindowWidth,
			wantHeight:    BrowseWindowHeight,
		},
		{
			name:          "mid sized monitor scales up beyond default",
			monitorWidth:  1800,
			monitorHeight: 1000,
			wantWidth:     1600,
			wantHeight:    900,
		},
		{
			name:          "common laptop scales above default when there is room",
			monitorWidth:  1440,
			monitorHeight: 900,
			wantWidth:     1296,
			wantHeight:    729,
		},
		{
			name:          "common laptop browse mode caps height",
			monitorWidth:  1440,
			monitorHeight: 900,
			browseMode:    true,
			wantWidth:     BrowseWindowWidth,
			wantHeight:    BrowseWindowHeight,
		},
		{
			name:          "small monitor scales down to fit",
			monitorWidth:  1280,
			monitorHeight: 720,
			wantWidth:     1152,
			wantHeight:    648,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWidth, gotHeight := InitialWindowSizeForMonitor(tt.monitorWidth, tt.monitorHeight, tt.browseMode)
			if gotWidth != tt.wantWidth || gotHeight != tt.wantHeight {
				t.Fatalf("InitialWindowSizeForMonitor(%d, %d, %t) = (%d, %d), want (%d, %d)", tt.monitorWidth, tt.monitorHeight, tt.browseMode, gotWidth, gotHeight, tt.wantWidth, tt.wantHeight)
			}
		})
	}
}
