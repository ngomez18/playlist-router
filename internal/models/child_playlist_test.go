package models

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildChildPlaylistName(t *testing.T) {
	tests := []struct {
		name              string
		basePlaylistName  string
		childPlaylistName string
		expectedResult    string
		expectedContains  []string
	}{
		{
			name:              "standard names",
			basePlaylistName:  "My Favorites",
			childPlaylistName: "High Energy",
			expectedResult:    "[My Favorites] > High Energy",
			expectedContains:  []string{"[My Favorites]", ">", "High Energy"},
		},
		{
			name:              "names with brackets and arrows",
			basePlaylistName:  "Base [with] brackets",
			childPlaylistName: "Child > with > arrows",
			expectedResult:    "[Base [with] brackets] > Child > with > arrows",
			expectedContains:  []string{"[Base [with] brackets]", ">", "Child > with > arrows"},
		},
		{
			name:              "empty child name",
			basePlaylistName:  "Base Playlist",
			childPlaylistName: "",
			expectedResult:    "[Base Playlist] > ",
			expectedContains:  []string{"[Base Playlist]", ">"},
		},
		{
			name:              "empty base name",
			basePlaylistName:  "",
			childPlaylistName: "Child Playlist",
			expectedResult:    "[] > Child Playlist",
			expectedContains:  []string{"[]", ">", "Child Playlist"},
		},
		{
			name:              "both names empty",
			basePlaylistName:  "",
			childPlaylistName: "",
			expectedResult:    "[] > ",
			expectedContains:  []string{"[]", ">"},
		},
		{
			name:              "names with unicode characters",
			basePlaylistName:  "Música Latina 🎵",
			childPlaylistName: "Reggaetón 💃",
			expectedResult:    "[Música Latina 🎵] > Reggaetón 💃",
			expectedContains:  []string{"[Música Latina 🎵]", ">", "Reggaetón 💃"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			result := BuildChildPlaylistName(tt.basePlaylistName, tt.childPlaylistName)

			// Verify exact result
			assert.Equal(tt.expectedResult, result)

			// Verify contains expected components
			for _, expectedPart := range tt.expectedContains {
				assert.Contains(result, expectedPart, "Result should contain: %s", expectedPart)
			}
		})
	}
}

func TestBuildChildPlaylistDescription(t *testing.T) {
	tests := []struct {
		name                string
		inputDescription    string
		expectedDescription string
	}{
		{
			name:             "with user description",
			inputDescription: "This is my custom description for the child playlist",
		},
		{
			name:             "empty description",
			inputDescription: "",
		},
		{
			name:             "description with special characters",
			inputDescription: "Special chars: @#$%^&*()_+{}|:<>?[]\\;'\",./ and unicode: 🎵🎶💃🕺",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert := require.New(t)

			result := BuildChildPlaylistDescription(tt.inputDescription)

			// Verify the result is not empty
			assert.NotEmpty(result, "Result should never be empty")

			// Verify expected content is present
			expectedDescription := fmt.Sprintf("[PLAYLIST GENERATED AND MANAGEED BY PlaylistRouter] %s", tt.inputDescription)
			assert.Equal(expectedDescription, result)

		})
	}
}
